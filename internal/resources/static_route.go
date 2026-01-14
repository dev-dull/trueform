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
	_ resource.Resource                = &StaticRouteResource{}
	_ resource.ResourceWithImportState = &StaticRouteResource{}
)

func NewStaticRouteResource() resource.Resource {
	return &StaticRouteResource{}
}

type StaticRouteResource struct {
	client *client.Client
}

type StaticRouteResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Destination types.String `tfsdk:"destination"`
	Gateway     types.String `tfsdk:"gateway"`
	Description types.String `tfsdk:"description"`
}

func (r *StaticRouteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_route"
}

func (r *StaticRouteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a static route on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier for the static route.",
				Computed:    true,
			},
			"destination": schema.StringAttribute{
				Description: "Destination network in CIDR notation (e.g., 10.0.0.0/8).",
				Required:    true,
			},
			"gateway": schema.StringAttribute{
				Description: "Gateway IP address.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the route.",
				Optional:    true,
			},
		},
	}
}

func (r *StaticRouteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StaticRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StaticRouteResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating static route", map[string]interface{}{
		"destination": plan.Destination.ValueString(),
		"gateway":     plan.Gateway.ValueString(),
	})

	createData := map[string]interface{}{
		"destination": plan.Destination.ValueString(),
		"gateway":     plan.Gateway.ValueString(),
	}

	if !plan.Description.IsNull() {
		createData["description"] = plan.Description.ValueString()
	}

	var result map[string]interface{}
	err := r.client.Create(ctx, "staticroute", createData, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Static Route", "Could not create static route: "+err.Error())
		return
	}

	routeID := int64(result["id"].(float64))
	if err := r.readStaticRoute(ctx, routeID, &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading Static Route", "Could not read static route after creation: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *StaticRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StaticRouteResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readStaticRoute(ctx, state.ID.ValueInt64(), &state); err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Static Route", "Could not read static route: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *StaticRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StaticRouteResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state StaticRouteResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateData := map[string]interface{}{}

	if !plan.Destination.Equal(state.Destination) {
		updateData["destination"] = plan.Destination.ValueString()
	}
	if !plan.Gateway.Equal(state.Gateway) {
		updateData["gateway"] = plan.Gateway.ValueString()
	}
	if !plan.Description.Equal(state.Description) {
		if plan.Description.IsNull() {
			updateData["description"] = ""
		} else {
			updateData["description"] = plan.Description.ValueString()
		}
	}

	if len(updateData) > 0 {
		var result map[string]interface{}
		err := r.client.Update(ctx, "staticroute", state.ID.ValueInt64(), updateData, &result)
		if err != nil {
			resp.Diagnostics.AddError("Error Updating Static Route", "Could not update static route: "+err.Error())
			return
		}
	}

	if err := r.readStaticRoute(ctx, state.ID.ValueInt64(), &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading Static Route", "Could not read static route after update: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *StaticRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StaticRouteResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "staticroute", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Static Route", "Could not delete static route: "+err.Error())
		return
	}
}

func (r *StaticRouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *StaticRouteResource) readStaticRoute(ctx context.Context, id int64, model *StaticRouteResourceModel) error {
	var result map[string]interface{}
	err := r.client.GetInstance(ctx, "staticroute", id, &result)
	if err != nil {
		return err
	}

	model.ID = types.Int64Value(int64(result["id"].(float64)))
	model.Destination = types.StringValue(result["destination"].(string))
	model.Gateway = types.StringValue(result["gateway"].(string))

	if description, ok := result["description"].(string); ok {
		model.Description = types.StringValue(description)
	}

	return nil
}
