package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/trueform/terraform-provider-trueform/internal/client"
)

var (
	_ resource.Resource                = &ISCSITargetResource{}
	_ resource.ResourceWithImportState = &ISCSITargetResource{}
)

func NewISCSITargetResource() resource.Resource {
	return &ISCSITargetResource{}
}

type ISCSITargetResource struct {
	client *client.Client
}

type ISCSITargetResourceModel struct {
	ID     types.Int64  `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Alias  types.String `tfsdk:"alias"`
	Mode   types.String `tfsdk:"mode"`
	Groups types.List   `tfsdk:"groups"`
}

type TargetGroup struct {
	Portal         types.Int64  `tfsdk:"portal"`
	Initiator      types.Int64  `tfsdk:"initiator"`
	AuthMethod     types.String `tfsdk:"authmethod"`
	Auth           types.Int64  `tfsdk:"auth"`
}

func (r *ISCSITargetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iscsi_target"
}

func (r *ISCSITargetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an iSCSI target on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier for the target.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The base name of the target (will be prefixed with IQN).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"alias": schema.StringAttribute{
				Description: "Optional alias for the target.",
				Optional:    true,
			},
			"mode": schema.StringAttribute{
				Description: "Target mode (ISCSI, FC, BOTH).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ISCSI"),
			},
			"groups": schema.ListNestedAttribute{
				Description: "List of portal groups for the target.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"portal": schema.Int64Attribute{
							Description: "Portal ID to use.",
							Required:    true,
						},
						"initiator": schema.Int64Attribute{
							Description: "Initiator group ID.",
							Optional:    true,
						},
						"authmethod": schema.StringAttribute{
							Description: "Authentication method (NONE, CHAP, CHAP_MUTUAL).",
							Optional:    true,
						},
						"auth": schema.Int64Attribute{
							Description: "Auth credential group ID.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (r *ISCSITargetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ISCSITargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ISCSITargetResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating iSCSI target", map[string]interface{}{
		"name": plan.Name.ValueString(),
	})

	createData := map[string]interface{}{
		"name": plan.Name.ValueString(),
		"mode": plan.Mode.ValueString(),
	}

	if !plan.Alias.IsNull() {
		createData["alias"] = plan.Alias.ValueString()
	}

	if !plan.Groups.IsNull() {
		var groupItems []TargetGroup
		diags = plan.Groups.ElementsAs(ctx, &groupItems, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		groups := make([]map[string]interface{}, len(groupItems))
		for i, item := range groupItems {
			groups[i] = map[string]interface{}{
				"portal": item.Portal.ValueInt64(),
			}
			if !item.Initiator.IsNull() {
				groups[i]["initiator"] = item.Initiator.ValueInt64()
			}
			if !item.AuthMethod.IsNull() {
				groups[i]["authmethod"] = item.AuthMethod.ValueString()
			}
			if !item.Auth.IsNull() {
				groups[i]["auth"] = item.Auth.ValueInt64()
			}
		}
		createData["groups"] = groups
	}

	var result map[string]interface{}
	err := r.client.Create(ctx, "iscsi.target", createData, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating iSCSI Target", "Could not create iSCSI target: "+err.Error())
		return
	}

	targetID := int64(result["id"].(float64))
	if err := r.readTarget(ctx, targetID, &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading iSCSI Target", "Could not read iSCSI target after creation: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ISCSITargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ISCSITargetResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readTarget(ctx, state.ID.ValueInt64(), &state); err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading iSCSI Target", "Could not read iSCSI target: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ISCSITargetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ISCSITargetResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ISCSITargetResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateData := map[string]interface{}{}

	if !plan.Alias.Equal(state.Alias) {
		if plan.Alias.IsNull() {
			updateData["alias"] = ""
		} else {
			updateData["alias"] = plan.Alias.ValueString()
		}
	}
	if !plan.Mode.Equal(state.Mode) {
		updateData["mode"] = plan.Mode.ValueString()
	}

	if !plan.Groups.Equal(state.Groups) {
		var groupItems []TargetGroup
		if !plan.Groups.IsNull() {
			diags = plan.Groups.ElementsAs(ctx, &groupItems, false)
			resp.Diagnostics.Append(diags...)
		}

		groups := make([]map[string]interface{}, len(groupItems))
		for i, item := range groupItems {
			groups[i] = map[string]interface{}{
				"portal": item.Portal.ValueInt64(),
			}
			if !item.Initiator.IsNull() {
				groups[i]["initiator"] = item.Initiator.ValueInt64()
			}
			if !item.AuthMethod.IsNull() {
				groups[i]["authmethod"] = item.AuthMethod.ValueString()
			}
			if !item.Auth.IsNull() {
				groups[i]["auth"] = item.Auth.ValueInt64()
			}
		}
		updateData["groups"] = groups
	}

	if len(updateData) > 0 {
		var result map[string]interface{}
		err := r.client.Update(ctx, "iscsi.target", state.ID.ValueInt64(), updateData, &result)
		if err != nil {
			resp.Diagnostics.AddError("Error Updating iSCSI Target", "Could not update iSCSI target: "+err.Error())
			return
		}
	}

	if err := r.readTarget(ctx, state.ID.ValueInt64(), &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading iSCSI Target", "Could not read iSCSI target after update: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ISCSITargetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ISCSITargetResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "iscsi.target", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting iSCSI Target", "Could not delete iSCSI target: "+err.Error())
		return
	}
}

func (r *ISCSITargetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *ISCSITargetResource) readTarget(ctx context.Context, id int64, model *ISCSITargetResourceModel) error {
	var result map[string]interface{}
	err := r.client.GetInstance(ctx, "iscsi.target", id, &result)
	if err != nil {
		return err
	}

	model.ID = types.Int64Value(int64(result["id"].(float64)))
	model.Name = types.StringValue(result["name"].(string))

	if alias, ok := result["alias"].(string); ok {
		model.Alias = types.StringValue(alias)
	}
	if mode, ok := result["mode"].(string); ok {
		model.Mode = types.StringValue(mode)
	}

	return nil
}
