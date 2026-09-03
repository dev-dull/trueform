package resources

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/trueform/terraform-provider-trueform/internal/client"
)

var (
	_ resource.Resource                = &SnapshotResource{}
	_ resource.ResourceWithImportState = &SnapshotResource{}
	_ resource.ResourceWithModifyPlan  = &SnapshotResource{}
)

func NewSnapshotResource() resource.Resource {
	return &SnapshotResource{}
}

type SnapshotResource struct {
	client *client.Client
}

type SnapshotResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Dataset            types.String `tfsdk:"dataset"`
	Name               types.String `tfsdk:"name"`
	Recursive          types.Bool   `tfsdk:"recursive"`
	VMWareSync         types.Bool   `tfsdk:"vmware_sync"`
	Properties         types.Map    `tfsdk:"properties"`
	Holds              types.List   `tfsdk:"holds"`
	ReferencedBytes    types.Int64  `tfsdk:"referenced_bytes"`
	UsedBytes          types.Int64  `tfsdk:"used_bytes"`
	CreationTime       types.String `tfsdk:"creation_time"`
}

func (r *SnapshotResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snapshot"
}

func (r *SnapshotResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a ZFS snapshot on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the snapshot (dataset@name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dataset": schema.StringAttribute{
				Description: "The dataset to snapshot (full path including pool).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the snapshot.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"recursive": schema.BoolAttribute{
				Description: "Create snapshots recursively for all child datasets.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"vmware_sync": schema.BoolAttribute{
				Description: "Sync with VMware before taking the snapshot.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"properties": schema.MapAttribute{
				Description: "Custom properties for the snapshot.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"holds": schema.ListAttribute{
				Description: "List of holds on the snapshot.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"referenced_bytes": schema.Int64Attribute{
				Description: "Amount of data referenced by the snapshot in bytes.",
				Computed:    true,
			},
			"used_bytes": schema.Int64Attribute{
				Description: "Amount of data used by the snapshot in bytes.",
				Computed:    true,
			},
			"creation_time": schema.StringAttribute{
				Description: "Creation time of the snapshot.",
				Computed:    true,
			},
		},
	}
}

func (r *SnapshotResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *SnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SnapshotResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	snapshotID := plan.Dataset.ValueString() + "@" + plan.Name.ValueString()

	tflog.Debug(ctx, "Creating snapshot", map[string]interface{}{
		"id": snapshotID,
	})

	var props map[string]string
	if !plan.Properties.IsNull() {
		diags = plan.Properties.ElementsAs(ctx, &props, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var err error
	if r.client.UsesZFSResourceSnapshot() {
		// TrueNAS 26+: zfs.resource.snapshot.create takes a single object and
		// only accepts user_properties at creation (no vmware_sync; regular ZFS
		// properties cannot be set at create time).
		createData := map[string]interface{}{
			"dataset":   plan.Dataset.ValueString(),
			"name":      plan.Name.ValueString(),
			"recursive": plan.Recursive.ValueBool(),
		}
		if props != nil {
			createData["user_properties"] = props
		}
		var entry map[string]interface{}
		err = r.client.Call(ctx, "zfs.resource.snapshot.create", []interface{}{createData}, &entry)
	} else {
		// TrueNAS 25.04–25.10: legacy zfs.snapshot.create.
		createData := map[string]interface{}{
			"dataset":   plan.Dataset.ValueString(),
			"name":      plan.Name.ValueString(),
			"recursive": plan.Recursive.ValueBool(),
		}
		if !plan.VMWareSync.IsNull() {
			createData["vmware_sync"] = plan.VMWareSync.ValueBool()
		}
		if props != nil {
			createData["properties"] = props
		}
		var result map[string]interface{}
		err = r.client.Create(ctx, "zfs.snapshot", createData, &result)
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Snapshot",
			"Could not create snapshot: "+err.Error(),
		)
		return
	}

	// Read the created snapshot
	if err := r.readSnapshot(ctx, snapshotID, &plan); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Snapshot",
			"Could not read snapshot after creation: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *SnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SnapshotResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readSnapshot(ctx, state.ID.ValueString(), &state); err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Snapshot",
			"Could not read snapshot: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

const snapshotPropsImmutableSummary = "Snapshot Properties Cannot Be Updated on TrueNAS 26+"

func snapshotPropsImmutableDetail(id string) string {
	return fmt.Sprintf(
		"TrueNAS 26 removed the snapshot update API, so the properties on %q cannot be changed in place. "+
			"Snapshot properties can only be set at creation time.\n\n"+
			"Either restore the previous \"properties\" value, or replace the snapshot with "+
			"\"terraform apply -replace=<resource address>\". Replacing destroys the snapshot and the "+
			"point-in-time data it holds.",
		id,
	)
}

// ModifyPlan rejects snapshot property changes against TrueNAS 26+ at plan time.
// TrueNAS 26 has no snapshot update method (the zfs.resource.snapshot.* namespace
// omits it, and there is no zfs.resource.update), so the change can never be
// applied. Update() repeats this check for the case where the client is not yet
// available here.
func (r *SnapshotResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// A null plan is a destroy and a null state is a create; neither is a property
	// update. A nil client means provider configuration was deferred, leaving the
	// server version unknown.
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() || r.client == nil {
		return
	}
	if !r.client.UsesZFSResourceSnapshot() {
		return
	}

	var plan, state SnapshotResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Properties.Equal(state.Properties) && !plan.Properties.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("properties"),
			snapshotPropsImmutableSummary,
			snapshotPropsImmutableDetail(state.ID.ValueString()),
		)
	}
}

func (r *SnapshotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SnapshotResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state SnapshotResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Snapshots have very limited update capabilities; only properties may change.
	if !plan.Properties.Equal(state.Properties) && !plan.Properties.IsNull() {
		if r.client.UsesZFSResourceSnapshot() {
			// Normally unreachable: ModifyPlan rejects this at plan time. This
			// backstops the case where the client was unavailable there. Failing
			// is required, not merely preferred -- Read() does not refresh
			// "properties" from the server, so a silent skip would write the
			// planned value to state, leave ZFS untouched, and produce no drift
			// on subsequent plans to reveal the discrepancy.
			resp.Diagnostics.AddAttributeError(
				path.Root("properties"),
				snapshotPropsImmutableSummary,
				snapshotPropsImmutableDetail(state.ID.ValueString()),
			)
			return
		}

		var props map[string]string
		diags = plan.Properties.ElementsAs(ctx, &props, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		updateData := map[string]interface{}{
			"user_properties_update": props,
		}

		var result map[string]interface{}
		err := r.client.Update(ctx, "zfs.snapshot", state.ID.ValueString(), updateData, &result)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Snapshot",
				"Could not update snapshot: "+err.Error(),
			)
			return
		}
	}

	// Read the updated snapshot
	if err := r.readSnapshot(ctx, state.ID.ValueString(), &plan); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Snapshot",
			"Could not read snapshot after update: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *SnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SnapshotResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting snapshot", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	var err error
	if r.client.UsesZFSResourceSnapshot() {
		// TrueNAS 26+: zfs.resource.snapshot.destroy takes a single object keyed
		// by "path" and returns null.
		destroyData := map[string]interface{}{
			"path":      state.ID.ValueString(),
			"recursive": state.Recursive.ValueBool(),
		}
		var res interface{}
		err = r.client.Call(ctx, "zfs.resource.snapshot.destroy", []interface{}{destroyData}, &res)
	} else {
		// TrueNAS 25.04–25.10: legacy zfs.snapshot.delete.
		deleteOptions := map[string]interface{}{
			"recursive": state.Recursive.ValueBool(),
		}
		err = r.client.DeleteWithOptions(ctx, "zfs.snapshot", state.ID.ValueString(), deleteOptions)
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Snapshot",
			"Could not delete snapshot: "+err.Error(),
		)
		return
	}
}

func (r *SnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *SnapshotResource) readSnapshot(ctx context.Context, id string, model *SnapshotResourceModel) error {
	if r.client.UsesZFSResourceSnapshot() {
		return r.readSnapshotV26(ctx, id, model)
	}
	return r.readSnapshotLegacy(ctx, id, model)
}

// readSnapshotLegacy reads a snapshot via the TrueNAS 25.04–25.10
// zfs.snapshot.get_instance API.
func (r *SnapshotResource) readSnapshotLegacy(ctx context.Context, id string, model *SnapshotResourceModel) error {
	var result map[string]interface{}
	if err := r.client.GetInstance(ctx, "zfs.snapshot", id, &result); err != nil {
		return err
	}

	model.ID = types.StringValue(result["id"].(string))
	r.setDatasetAndName(id, model)
	r.setHolds(ctx, result["holds"], model)
	r.setWriteOnlyDefaults(model)

	if properties, ok := result["properties"].(map[string]interface{}); ok {
		model.ReferencedBytes = propInt64(properties, "referenced", "parsed")
		model.UsedBytes = propInt64(properties, "used", "parsed")
		if creation, ok := properties["creation"].(map[string]interface{}); ok {
			// creation.parsed can be either a string or a timestamp
			if parsed, ok := creation["parsed"].(string); ok {
				model.CreationTime = types.StringValue(parsed)
			} else if rawValue, ok := creation["rawvalue"].(string); ok {
				model.CreationTime = types.StringValue(rawValue)
			} else if value, ok := creation["value"].(string); ok {
				model.CreationTime = types.StringValue(value)
			} else {
				model.CreationTime = types.StringNull()
			}
		} else {
			model.CreationTime = types.StringNull()
		}
	} else {
		model.CreationTime = types.StringNull()
	}

	return nil
}

// readSnapshotV26 reads a snapshot via the TrueNAS 26+ zfs.resource.snapshot.query
// API. The response shape differs from the legacy API: the identifier is "name"
// (not "id"), and property values live under ".value" (not ".parsed"), with
// creation expressed as a unix-epoch number.
func (r *SnapshotResource) readSnapshotV26(ctx context.Context, id string, model *SnapshotResourceModel) error {
	args := map[string]interface{}{
		"paths":               []string{id},
		"properties":          []string{"referenced", "used", "creation"},
		"get_user_properties": true,
		"get_holds":           true,
	}
	var entries []map[string]interface{}
	if err := r.client.Call(ctx, "zfs.resource.snapshot.query", []interface{}{args}, &entries); err != nil {
		return err
	}
	if len(entries) == 0 {
		return &client.APIError{Code: client.ErrCodeNotFound, Message: "snapshot " + id + " not found"}
	}
	result := entries[0]

	if name, ok := result["name"].(string); ok {
		model.ID = types.StringValue(name)
	} else {
		model.ID = types.StringValue(id)
	}
	r.setDatasetAndName(id, model)
	r.setHolds(ctx, result["holds"], model)
	r.setWriteOnlyDefaults(model)

	if properties, ok := result["properties"].(map[string]interface{}); ok {
		model.ReferencedBytes = propInt64(properties, "referenced", "value")
		model.UsedBytes = propInt64(properties, "used", "value")
		if creation, ok := properties["creation"].(map[string]interface{}); ok {
			if epoch, ok := creation["value"].(float64); ok {
				model.CreationTime = types.StringValue(time.Unix(int64(epoch), 0).UTC().Format(time.RFC3339))
			} else if raw, ok := creation["raw"].(string); ok {
				model.CreationTime = types.StringValue(raw)
			} else {
				model.CreationTime = types.StringNull()
			}
		} else {
			model.CreationTime = types.StringNull()
		}
	} else {
		model.CreationTime = types.StringNull()
	}

	return nil
}

// setDatasetAndName splits the snapshot id (dataset@name) into its components.
func (r *SnapshotResource) setDatasetAndName(id string, model *SnapshotResourceModel) {
	parts := strings.SplitN(id, "@", 2)
	if len(parts) == 2 {
		model.Dataset = types.StringValue(parts[0])
		model.Name = types.StringValue(parts[1])
	}
}

// setHolds populates the computed holds list. The legacy API returns a list of
// strings; the 26+ API returns a list of objects (hold tags). Both are reduced
// to a list of tag names; an absent/empty value yields an empty list.
func (r *SnapshotResource) setHolds(ctx context.Context, raw interface{}, model *SnapshotResourceModel) {
	var holdsList []string
	if holds, ok := raw.([]interface{}); ok {
		for _, h := range holds {
			switch v := h.(type) {
			case string:
				holdsList = append(holdsList, v)
			case map[string]interface{}:
				if tag, ok := v["tag"].(string); ok {
					holdsList = append(holdsList, tag)
				} else if name, ok := v["name"].(string); ok {
					holdsList = append(holdsList, name)
				}
			}
		}
	}
	holdValues, diags := types.ListValueFrom(ctx, types.StringType, holdsList)
	if !diags.HasError() {
		model.Holds = holdValues
	}
}

// setWriteOnlyDefaults defaults the write-only flags (recursive, vmware_sync)
// that the snapshot record does not echo back. On a fresh import the model is
// empty, so these default to false to match the schema; on refresh of an
// already-tracked resource the prior non-null state value passes through.
func (r *SnapshotResource) setWriteOnlyDefaults(model *SnapshotResourceModel) {
	if model.Recursive.IsNull() || model.Recursive.IsUnknown() {
		model.Recursive = types.BoolValue(false)
	}
	if model.VMWareSync.IsNull() || model.VMWareSync.IsUnknown() {
		model.VMWareSync = types.BoolValue(false)
	}
}

// propInt64 reads properties[key][field] as an int64, returning null if absent.
func propInt64(properties map[string]interface{}, key, field string) types.Int64 {
	if sub, ok := properties[key].(map[string]interface{}); ok {
		if n, ok := sub[field].(float64); ok {
			return types.Int64Value(int64(n))
		}
	}
	return types.Int64Null()
}
