package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/trueform/terraform-provider-trueform/internal/client"
)

var (
	_ resource.Resource                = &ISCSIPortalResource{}
	_ resource.ResourceWithImportState = &ISCSIPortalResource{}
)

func NewISCSIPortalResource() resource.Resource {
	return &ISCSIPortalResource{}
}

type ISCSIPortalResource struct {
	client *client.Client
}

type ISCSIPortalResourceModel struct {
	ID            types.Int64  `tfsdk:"id"`
	Comment       types.String `tfsdk:"comment"`
	DiscoveryAuth types.String `tfsdk:"discovery_authmethod"`
	DiscoveryGroup types.Int64 `tfsdk:"discovery_authgroup"`
	Listen        types.List   `tfsdk:"listen"`
}

type PortalListen struct {
	IP   types.String `tfsdk:"ip"`
	Port types.Int64  `tfsdk:"port"`
}

func (r *ISCSIPortalResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iscsi_portal"
}

func (r *ISCSIPortalResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an iSCSI portal on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier for the portal.",
				Computed:    true,
			},
			"comment": schema.StringAttribute{
				Description: "Comment for the portal.",
				Optional:    true,
			},
			"discovery_authmethod": schema.StringAttribute{
				Description: "Discovery authentication method (NONE, CHAP, CHAP_MUTUAL).",
				Optional:    true,
			},
			"discovery_authgroup": schema.Int64Attribute{
				Description: "Discovery authentication group.",
				Optional:    true,
			},
			"listen": schema.ListNestedAttribute{
				Description: "List of IP addresses and ports to listen on.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip": schema.StringAttribute{
							Description: "IP address to listen on (0.0.0.0 for all).",
							Required:    true,
						},
						"port": schema.Int64Attribute{
							Description: "Port to listen on.",
							Optional:    true,
							Computed:    true,
							Default:     int64default.StaticInt64(3260),
						},
					},
				},
			},
		},
	}
}

func (r *ISCSIPortalResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ISCSIPortalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ISCSIPortalResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating iSCSI portal")

	var listenItems []PortalListen
	diags = plan.Listen.ElementsAs(ctx, &listenItems, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TrueNAS Scale 25 only accepts IP in listen configuration, port is implicit (3260)
	listen := make([]map[string]interface{}, len(listenItems))
	for i, item := range listenItems {
		listen[i] = map[string]interface{}{
			"ip": item.IP.ValueString(),
		}
	}

	createData := map[string]interface{}{
		"listen": listen,
	}

	if !plan.Comment.IsNull() {
		createData["comment"] = plan.Comment.ValueString()
	}
	if !plan.DiscoveryAuth.IsNull() {
		createData["discovery_authmethod"] = plan.DiscoveryAuth.ValueString()
	}
	if !plan.DiscoveryGroup.IsNull() {
		createData["discovery_authgroup"] = plan.DiscoveryGroup.ValueInt64()
	}

	var result map[string]interface{}
	err := r.client.Create(ctx, "iscsi.portal", createData, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating iSCSI Portal", "Could not create iSCSI portal: "+err.Error())
		return
	}

	portalID := int64(result["id"].(float64))
	if err := r.readPortal(ctx, portalID, &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading iSCSI Portal", "Could not read iSCSI portal after creation: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ISCSIPortalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ISCSIPortalResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readPortal(ctx, state.ID.ValueInt64(), &state); err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading iSCSI Portal", "Could not read iSCSI portal: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ISCSIPortalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ISCSIPortalResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ISCSIPortalResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var listenItems []PortalListen
	diags = plan.Listen.ElementsAs(ctx, &listenItems, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TrueNAS Scale 25 only accepts IP in listen configuration, port is implicit (3260)
	listen := make([]map[string]interface{}, len(listenItems))
	for i, item := range listenItems {
		listen[i] = map[string]interface{}{
			"ip": item.IP.ValueString(),
		}
	}

	updateData := map[string]interface{}{
		"listen": listen,
	}

	if !plan.Comment.IsNull() {
		updateData["comment"] = plan.Comment.ValueString()
	}
	if !plan.DiscoveryAuth.IsNull() {
		updateData["discovery_authmethod"] = plan.DiscoveryAuth.ValueString()
	}
	if !plan.DiscoveryGroup.IsNull() {
		updateData["discovery_authgroup"] = plan.DiscoveryGroup.ValueInt64()
	}

	var result map[string]interface{}
	err := r.client.Update(ctx, "iscsi.portal", state.ID.ValueInt64(), updateData, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating iSCSI Portal", "Could not update iSCSI portal: "+err.Error())
		return
	}

	if err := r.readPortal(ctx, state.ID.ValueInt64(), &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading iSCSI Portal", "Could not read iSCSI portal after update: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ISCSIPortalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ISCSIPortalResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "iscsi.portal", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting iSCSI Portal", "Could not delete iSCSI portal: "+err.Error())
		return
	}
}

func (r *ISCSIPortalResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *ISCSIPortalResource) readPortal(ctx context.Context, id int64, model *ISCSIPortalResourceModel) error {
	var result map[string]interface{}
	err := r.client.GetInstance(ctx, "iscsi.portal", id, &result)
	if err != nil {
		return err
	}

	model.ID = types.Int64Value(int64(result["id"].(float64)))
	if comment, ok := result["comment"].(string); ok {
		model.Comment = types.StringValue(comment)
	}
	if discoveryAuth, ok := result["discovery_authmethod"].(string); ok {
		model.DiscoveryAuth = types.StringValue(discoveryAuth)
	}
	if discoveryGroup, ok := result["discovery_authgroup"].(float64); ok {
		model.DiscoveryGroup = types.Int64Value(int64(discoveryGroup))
	}

	if listenList, ok := result["listen"].([]interface{}); ok {
		listenItems := make([]PortalListen, len(listenList))
		for i, item := range listenList {
			if listenMap, ok := item.(map[string]interface{}); ok {
				listenItems[i] = PortalListen{
					IP:   types.StringValue(listenMap["ip"].(string)),
					Port: types.Int64Value(int64(listenMap["port"].(float64))),
				}
			}
		}
		listenValue, d := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"ip":   types.StringType,
				"port": types.Int64Type,
			},
		}, listenItems)
		if !d.HasError() {
			model.Listen = listenValue
		}
	}

	return nil
}
