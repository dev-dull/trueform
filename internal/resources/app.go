package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/trueform/terraform-provider-trueform/internal/client"
)

var (
	_ resource.Resource                = &AppResource{}
	_ resource.ResourceWithImportState = &AppResource{}
)

func NewAppResource() resource.Resource {
	return &AppResource{}
}

type AppResource struct {
	client *client.Client
}

type AppResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	CatalogApp  types.String `tfsdk:"catalog_app"`
	Train       types.String `tfsdk:"train"`
	Version     types.String `tfsdk:"version"`
	Values      types.String `tfsdk:"values"`
	State       types.String `tfsdk:"state"`
	Metadata    types.Map    `tfsdk:"metadata"`
}

func (r *AppResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (r *AppResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an application on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the app (same as name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the app instance.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"catalog_app": schema.StringAttribute{
				Description: "The catalog app to deploy (e.g., 'plex', 'nextcloud').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"train": schema.StringAttribute{
				Description: "The catalog train (e.g., 'stable', 'community').",
				Optional:    true,
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "The app version to deploy.",
				Optional:    true,
				Computed:    true,
			},
			"values": schema.StringAttribute{
				Description: "JSON-encoded configuration values for the app.",
				Optional:    true,
			},
			"state": schema.StringAttribute{
				Description: "Current state of the app.",
				Computed:    true,
			},
			"metadata": schema.MapAttribute{
				Description: "App metadata.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *AppResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AppResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating app", map[string]interface{}{
		"name":        plan.Name.ValueString(),
		"catalog_app": plan.CatalogApp.ValueString(),
	})

	createData := map[string]interface{}{
		"app_name":    plan.Name.ValueString(),
		"catalog_app": plan.CatalogApp.ValueString(),
	}

	if !plan.Train.IsNull() {
		createData["train"] = plan.Train.ValueString()
	}
	if !plan.Version.IsNull() {
		createData["version"] = plan.Version.ValueString()
	}

	if !plan.Values.IsNull() && plan.Values.ValueString() != "" {
		var values map[string]interface{}
		if err := json.Unmarshal([]byte(plan.Values.ValueString()), &values); err != nil {
			resp.Diagnostics.AddError("Invalid Values JSON", "Could not parse values JSON: "+err.Error())
			return
		}
		createData["values"] = values
	}

	var result map[string]interface{}
	err := r.client.Create(ctx, "app", createData, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating App", "Could not create app: "+err.Error())
		return
	}

	if err := r.readApp(ctx, plan.Name.ValueString(), &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading App", "Could not read app after creation: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *AppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AppResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readApp(ctx, state.ID.ValueString(), &state); err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading App", "Could not read app: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *AppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AppResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AppResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating app", map[string]interface{}{
		"name": state.ID.ValueString(),
	})

	updateData := map[string]interface{}{}

	if !plan.Values.Equal(state.Values) && !plan.Values.IsNull() {
		var values map[string]interface{}
		if err := json.Unmarshal([]byte(plan.Values.ValueString()), &values); err != nil {
			resp.Diagnostics.AddError("Invalid Values JSON", "Could not parse values JSON: "+err.Error())
			return
		}
		updateData["values"] = values
	}

	if len(updateData) > 0 {
		var result map[string]interface{}
		err := r.client.Update(ctx, "app", state.ID.ValueString(), updateData, &result)
		if err != nil {
			resp.Diagnostics.AddError("Error Updating App", "Could not update app: "+err.Error())
			return
		}
	}

	// Handle version upgrade
	if !plan.Version.Equal(state.Version) && !plan.Version.IsNull() {
		err := r.client.Call(ctx, "app.upgrade", []interface{}{
			state.ID.ValueString(),
			map[string]interface{}{
				"app_version": plan.Version.ValueString(),
			},
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Error Upgrading App", "Could not upgrade app: "+err.Error())
			return
		}
	}

	if err := r.readApp(ctx, state.ID.ValueString(), &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading App", "Could not read app after update: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *AppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AppResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting app", map[string]interface{}{
		"name": state.ID.ValueString(),
	})

	err := r.client.Delete(ctx, "app", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting App", "Could not delete app: "+err.Error())
		return
	}
}

func (r *AppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *AppResource) readApp(ctx context.Context, name string, model *AppResourceModel) error {
	var result map[string]interface{}
	err := r.client.GetInstance(ctx, "app", name, &result)
	if err != nil {
		return err
	}

	model.ID = types.StringValue(result["id"].(string))
	model.Name = types.StringValue(result["name"].(string))

	if metadata, ok := result["metadata"].(map[string]interface{}); ok {
		if appVersion, ok := metadata["app_version"].(string); ok {
			model.Version = types.StringValue(appVersion)
		}
		if train, ok := metadata["train"].(string); ok {
			model.Train = types.StringValue(train)
		}
	}

	if state, ok := result["state"].(string); ok {
		model.State = types.StringValue(state)
	}

	return nil
}
