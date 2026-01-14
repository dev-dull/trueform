package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/trueform/terraform-provider-trueform/internal/client"
)

var (
	_ resource.Resource                = &ISCSITargetExtentResource{}
	_ resource.ResourceWithImportState = &ISCSITargetExtentResource{}
)

func NewISCSITargetExtentResource() resource.Resource {
	return &ISCSITargetExtentResource{}
}

type ISCSITargetExtentResource struct {
	client *client.Client
}

type ISCSITargetExtentResourceModel struct {
	ID     types.Int64 `tfsdk:"id"`
	Target types.Int64 `tfsdk:"target"`
	Extent types.Int64 `tfsdk:"extent"`
	LunID  types.Int64 `tfsdk:"lunid"`
}

func (r *ISCSITargetExtentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iscsi_targetextent"
}

func (r *ISCSITargetExtentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an iSCSI target-to-extent (LUN) mapping on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier for the target-extent mapping.",
				Computed:    true,
			},
			"target": schema.Int64Attribute{
				Description: "The target ID to map.",
				Required:    true,
			},
			"extent": schema.Int64Attribute{
				Description: "The extent ID to map.",
				Required:    true,
			},
			"lunid": schema.Int64Attribute{
				Description: "The LUN ID for this mapping (0-1023).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
		},
	}
}

func (r *ISCSITargetExtentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData))
		return
	}
	r.client = client
}

func (r *ISCSITargetExtentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ISCSITargetExtentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating iSCSI target-extent mapping", map[string]interface{}{
		"target": plan.Target.ValueInt64(),
		"extent": plan.Extent.ValueInt64(),
	})

	createData := map[string]interface{}{
		"target": plan.Target.ValueInt64(),
		"extent": plan.Extent.ValueInt64(),
		"lunid":  plan.LunID.ValueInt64(),
	}

	var result map[string]interface{}
	err := r.client.Create(ctx, "iscsi.targetextent", createData, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating iSCSI Target-Extent", "Could not create iSCSI target-extent mapping: "+err.Error())
		return
	}

	mappingID := int64(result["id"].(float64))
	if err := r.readTargetExtent(ctx, mappingID, &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading iSCSI Target-Extent", "Could not read iSCSI target-extent mapping after creation: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ISCSITargetExtentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ISCSITargetExtentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readTargetExtent(ctx, state.ID.ValueInt64(), &state); err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading iSCSI Target-Extent", "Could not read iSCSI target-extent mapping: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ISCSITargetExtentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ISCSITargetExtentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ISCSITargetExtentResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateData := map[string]interface{}{}

	if !plan.Target.Equal(state.Target) {
		updateData["target"] = plan.Target.ValueInt64()
	}
	if !plan.Extent.Equal(state.Extent) {
		updateData["extent"] = plan.Extent.ValueInt64()
	}
	if !plan.LunID.Equal(state.LunID) {
		updateData["lunid"] = plan.LunID.ValueInt64()
	}

	if len(updateData) > 0 {
		var result map[string]interface{}
		err := r.client.Update(ctx, "iscsi.targetextent", state.ID.ValueInt64(), updateData, &result)
		if err != nil {
			resp.Diagnostics.AddError("Error Updating iSCSI Target-Extent", "Could not update iSCSI target-extent mapping: "+err.Error())
			return
		}
	}

	if err := r.readTargetExtent(ctx, state.ID.ValueInt64(), &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading iSCSI Target-Extent", "Could not read iSCSI target-extent mapping after update: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ISCSITargetExtentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ISCSITargetExtentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "iscsi.targetextent", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting iSCSI Target-Extent", "Could not delete iSCSI target-extent mapping: "+err.Error())
		return
	}
}

func (r *ISCSITargetExtentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Could not parse import ID %q as integer: %v", req.ID, err),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func (r *ISCSITargetExtentResource) readTargetExtent(ctx context.Context, id int64, model *ISCSITargetExtentResourceModel) error {
	var result map[string]interface{}
	err := r.client.GetInstance(ctx, "iscsi.targetextent", id, &result)
	if err != nil {
		return err
	}

	model.ID = types.Int64Value(int64(result["id"].(float64)))
	model.Target = types.Int64Value(int64(result["target"].(float64)))
	model.Extent = types.Int64Value(int64(result["extent"].(float64)))
	model.LunID = types.Int64Value(int64(result["lunid"].(float64)))

	return nil
}
