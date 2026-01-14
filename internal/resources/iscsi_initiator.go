package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/trueform/terraform-provider-trueform/internal/client"
)

var (
	_ resource.Resource                = &ISCSIInitiatorResource{}
	_ resource.ResourceWithImportState = &ISCSIInitiatorResource{}
)

func NewISCSIInitiatorResource() resource.Resource {
	return &ISCSIInitiatorResource{}
}

type ISCSIInitiatorResource struct {
	client *client.Client
}

type ISCSIInitiatorResourceModel struct {
	ID         types.Int64  `tfsdk:"id"`
	Comment    types.String `tfsdk:"comment"`
	Initiators types.List   `tfsdk:"initiators"`
	AuthNetwork types.List  `tfsdk:"auth_network"`
}

func (r *ISCSIInitiatorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iscsi_initiator"
}

func (r *ISCSIInitiatorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an iSCSI initiator group on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier for the initiator group.",
				Computed:    true,
			},
			"comment": schema.StringAttribute{
				Description: "Comment for the initiator group.",
				Optional:    true,
			},
			"initiators": schema.ListAttribute{
				Description: "List of allowed initiator IQNs. Empty list allows all initiators.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"auth_network": schema.ListAttribute{
				Description: "List of allowed networks (CIDR notation). Empty list allows all networks.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *ISCSIInitiatorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ISCSIInitiatorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ISCSIInitiatorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating iSCSI initiator group")

	createData := map[string]interface{}{}

	if !plan.Comment.IsNull() {
		createData["comment"] = plan.Comment.ValueString()
	}
	if !plan.Initiators.IsNull() {
		var initiators []string
		diags = plan.Initiators.ElementsAs(ctx, &initiators, false)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			createData["initiators"] = initiators
		}
	}
	if !plan.AuthNetwork.IsNull() {
		var authNetwork []string
		diags = plan.AuthNetwork.ElementsAs(ctx, &authNetwork, false)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			createData["auth_network"] = authNetwork
		}
	}

	var result map[string]interface{}
	err := r.client.Create(ctx, "iscsi.initiator", createData, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating iSCSI Initiator", "Could not create iSCSI initiator: "+err.Error())
		return
	}

	initiatorID := int64(result["id"].(float64))
	if err := r.readInitiator(ctx, initiatorID, &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading iSCSI Initiator", "Could not read iSCSI initiator after creation: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ISCSIInitiatorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ISCSIInitiatorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readInitiator(ctx, state.ID.ValueInt64(), &state); err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading iSCSI Initiator", "Could not read iSCSI initiator: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ISCSIInitiatorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ISCSIInitiatorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ISCSIInitiatorResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateData := map[string]interface{}{}

	if !plan.Comment.Equal(state.Comment) {
		if plan.Comment.IsNull() {
			updateData["comment"] = ""
		} else {
			updateData["comment"] = plan.Comment.ValueString()
		}
	}
	if !plan.Initiators.Equal(state.Initiators) {
		var initiators []string
		if !plan.Initiators.IsNull() {
			diags = plan.Initiators.ElementsAs(ctx, &initiators, false)
			resp.Diagnostics.Append(diags...)
		}
		updateData["initiators"] = initiators
	}
	if !plan.AuthNetwork.Equal(state.AuthNetwork) {
		var authNetwork []string
		if !plan.AuthNetwork.IsNull() {
			diags = plan.AuthNetwork.ElementsAs(ctx, &authNetwork, false)
			resp.Diagnostics.Append(diags...)
		}
		updateData["auth_network"] = authNetwork
	}

	if len(updateData) > 0 {
		var result map[string]interface{}
		err := r.client.Update(ctx, "iscsi.initiator", state.ID.ValueInt64(), updateData, &result)
		if err != nil {
			resp.Diagnostics.AddError("Error Updating iSCSI Initiator", "Could not update iSCSI initiator: "+err.Error())
			return
		}
	}

	if err := r.readInitiator(ctx, state.ID.ValueInt64(), &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading iSCSI Initiator", "Could not read iSCSI initiator after update: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ISCSIInitiatorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ISCSIInitiatorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "iscsi.initiator", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting iSCSI Initiator", "Could not delete iSCSI initiator: "+err.Error())
		return
	}
}

func (r *ISCSIInitiatorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ISCSIInitiatorResource) readInitiator(ctx context.Context, id int64, model *ISCSIInitiatorResourceModel) error {
	var result map[string]interface{}
	err := r.client.GetInstance(ctx, "iscsi.initiator", id, &result)
	if err != nil {
		return err
	}

	model.ID = types.Int64Value(int64(result["id"].(float64)))

	if comment, ok := result["comment"].(string); ok {
		model.Comment = types.StringValue(comment)
	}
	if initiators, ok := result["initiators"].([]interface{}); ok {
		initiatorList := make([]string, len(initiators))
		for i, init := range initiators {
			initiatorList[i] = init.(string)
		}
		initiatorValues, diags := types.ListValueFrom(ctx, types.StringType, initiatorList)
		if !diags.HasError() {
			model.Initiators = initiatorValues
		}
	}
	if authNetwork, ok := result["auth_network"].([]interface{}); ok {
		networkList := make([]string, len(authNetwork))
		for i, net := range authNetwork {
			networkList[i] = net.(string)
		}
		networkValues, diags := types.ListValueFrom(ctx, types.StringType, networkList)
		if !diags.HasError() {
			model.AuthNetwork = networkValues
		}
	}

	return nil
}
