package resources

import (
	"context"
	"fmt"
	"strings"

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

	createData := map[string]interface{}{
		"dataset":   plan.Dataset.ValueString(),
		"name":      plan.Name.ValueString(),
		"recursive": plan.Recursive.ValueBool(),
	}

	if !plan.VMWareSync.IsNull() {
		createData["vmware_sync"] = plan.VMWareSync.ValueBool()
	}

	if !plan.Properties.IsNull() {
		var props map[string]string
		diags = plan.Properties.ElementsAs(ctx, &props, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		createData["properties"] = props
	}

	var result map[string]interface{}
	err := r.client.Create(ctx, "zfs.snapshot", createData, &result)
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

	// Snapshots have very limited update capabilities
	// Properties might be updatable
	if !plan.Properties.Equal(state.Properties) && !plan.Properties.IsNull() {
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

	deleteOptions := map[string]interface{}{
		"recursive": state.Recursive.ValueBool(),
	}

	err := r.client.DeleteWithOptions(ctx, "zfs.snapshot", state.ID.ValueString(), deleteOptions)
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
	var result map[string]interface{}
	err := r.client.GetInstance(ctx, "zfs.snapshot", id, &result)
	if err != nil {
		return err
	}

	model.ID = types.StringValue(result["id"].(string))

	// Parse dataset and name from the ID
	parts := strings.SplitN(id, "@", 2)
	if len(parts) == 2 {
		model.Dataset = types.StringValue(parts[0])
		model.Name = types.StringValue(parts[1])
	}

	if holds, ok := result["holds"].([]interface{}); ok && len(holds) > 0 {
		holdsList := make([]string, len(holds))
		for i, h := range holds {
			holdsList[i] = h.(string)
		}
		holdValues, diags := types.ListValueFrom(ctx, types.StringType, holdsList)
		if !diags.HasError() {
			model.Holds = holdValues
		}
	} else {
		// Set empty list when no holds
		emptyHolds, _ := types.ListValueFrom(ctx, types.StringType, []string{})
		model.Holds = emptyHolds
	}

	if properties, ok := result["properties"].(map[string]interface{}); ok {
		if referenced, ok := properties["referenced"].(map[string]interface{}); ok {
			if parsed, ok := referenced["parsed"].(float64); ok {
				model.ReferencedBytes = types.Int64Value(int64(parsed))
			}
		}
		if used, ok := properties["used"].(map[string]interface{}); ok {
			if parsed, ok := used["parsed"].(float64); ok {
				model.UsedBytes = types.Int64Value(int64(parsed))
			}
		}
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
