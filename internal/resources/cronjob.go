package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/trueform/terraform-provider-trueform/internal/client"
)

var (
	_ resource.Resource                = &CronjobResource{}
	_ resource.ResourceWithImportState = &CronjobResource{}
)

func NewCronjobResource() resource.Resource {
	return &CronjobResource{}
}

type CronjobResource struct {
	client *client.Client
}

type CronjobResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	User        types.String `tfsdk:"user"`
	Command     types.String `tfsdk:"command"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	StdOut      types.Bool   `tfsdk:"stdout"`
	StdErr      types.Bool   `tfsdk:"stderr"`
	Schedule    types.Object `tfsdk:"schedule"`
}

type CronSchedule struct {
	Minute  types.String `tfsdk:"minute"`
	Hour    types.String `tfsdk:"hour"`
	Dom     types.String `tfsdk:"dom"`
	Month   types.String `tfsdk:"month"`
	Dow     types.String `tfsdk:"dow"`
}

func (r *CronjobResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cronjob"
}

func (r *CronjobResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a cron job on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier for the cron job.",
				Computed:    true,
			},
			"user": schema.StringAttribute{
				Description: "The user to run the command as.",
				Required:    true,
			},
			"command": schema.StringAttribute{
				Description: "The command to execute.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the cron job.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the cron job is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"stdout": schema.BoolAttribute{
				Description: "Redirect stdout to syslog.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"stderr": schema.BoolAttribute{
				Description: "Redirect stderr to syslog.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"schedule": schema.SingleNestedAttribute{
				Description: "Cron schedule configuration.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"minute": schema.StringAttribute{
						Description: "Minute (0-59, or cron expression).",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("0"),
					},
					"hour": schema.StringAttribute{
						Description: "Hour (0-23, or cron expression).",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("0"),
					},
					"dom": schema.StringAttribute{
						Description: "Day of month (1-31, or cron expression).",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("*"),
					},
					"month": schema.StringAttribute{
						Description: "Month (1-12, or cron expression).",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("*"),
					},
					"dow": schema.StringAttribute{
						Description: "Day of week (0-6, or cron expression).",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("*"),
					},
				},
			},
		},
	}
}

func (r *CronjobResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CronjobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CronjobResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating cron job", map[string]interface{}{
		"user":    plan.User.ValueString(),
		"command": plan.Command.ValueString(),
	})

	var schedule CronSchedule
	diags = plan.Schedule.As(ctx, &schedule, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createData := map[string]interface{}{
		"user":    plan.User.ValueString(),
		"command": plan.Command.ValueString(),
		"enabled": plan.Enabled.ValueBool(),
		"stdout":  plan.StdOut.ValueBool(),
		"stderr":  plan.StdErr.ValueBool(),
		"schedule": map[string]interface{}{
			"minute": schedule.Minute.ValueString(),
			"hour":   schedule.Hour.ValueString(),
			"dom":    schedule.Dom.ValueString(),
			"month":  schedule.Month.ValueString(),
			"dow":    schedule.Dow.ValueString(),
		},
	}

	if !plan.Description.IsNull() {
		createData["description"] = plan.Description.ValueString()
	}

	var result map[string]interface{}
	err := r.client.Create(ctx, "cronjob", createData, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Cron Job", "Could not create cron job: "+err.Error())
		return
	}

	jobID := int64(result["id"].(float64))
	if err := r.readCronjob(ctx, jobID, &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading Cron Job", "Could not read cron job after creation: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CronjobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CronjobResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readCronjob(ctx, state.ID.ValueInt64(), &state); err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Cron Job", "Could not read cron job: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *CronjobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CronjobResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CronjobResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var schedule CronSchedule
	diags = plan.Schedule.As(ctx, &schedule, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateData := map[string]interface{}{
		"user":    plan.User.ValueString(),
		"command": plan.Command.ValueString(),
		"enabled": plan.Enabled.ValueBool(),
		"stdout":  plan.StdOut.ValueBool(),
		"stderr":  plan.StdErr.ValueBool(),
		"schedule": map[string]interface{}{
			"minute": schedule.Minute.ValueString(),
			"hour":   schedule.Hour.ValueString(),
			"dom":    schedule.Dom.ValueString(),
			"month":  schedule.Month.ValueString(),
			"dow":    schedule.Dow.ValueString(),
		},
	}

	if !plan.Description.IsNull() {
		updateData["description"] = plan.Description.ValueString()
	}

	var result map[string]interface{}
	err := r.client.Update(ctx, "cronjob", state.ID.ValueInt64(), updateData, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Cron Job", "Could not update cron job: "+err.Error())
		return
	}

	if err := r.readCronjob(ctx, state.ID.ValueInt64(), &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading Cron Job", "Could not read cron job after update: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CronjobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CronjobResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "cronjob", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Cron Job", "Could not delete cron job: "+err.Error())
		return
	}
}

func (r *CronjobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *CronjobResource) readCronjob(ctx context.Context, id int64, model *CronjobResourceModel) error {
	var result map[string]interface{}
	err := r.client.GetInstance(ctx, "cronjob", id, &result)
	if err != nil {
		return err
	}

	model.ID = types.Int64Value(int64(result["id"].(float64)))
	model.User = types.StringValue(result["user"].(string))
	model.Command = types.StringValue(result["command"].(string))

	if description, ok := result["description"].(string); ok {
		model.Description = types.StringValue(description)
	}
	if enabled, ok := result["enabled"].(bool); ok {
		model.Enabled = types.BoolValue(enabled)
	}
	if stdout, ok := result["stdout"].(bool); ok {
		model.StdOut = types.BoolValue(stdout)
	}
	if stderr, ok := result["stderr"].(bool); ok {
		model.StdErr = types.BoolValue(stderr)
	}

	if sched, ok := result["schedule"].(map[string]interface{}); ok {
		scheduleObj, d := types.ObjectValue(
			map[string]attr.Type{
				"minute": types.StringType,
				"hour":   types.StringType,
				"dom":    types.StringType,
				"month":  types.StringType,
				"dow":    types.StringType,
			},
			map[string]attr.Value{
				"minute": types.StringValue(sched["minute"].(string)),
				"hour":   types.StringValue(sched["hour"].(string)),
				"dom":    types.StringValue(sched["dom"].(string)),
				"month":  types.StringValue(sched["month"].(string)),
				"dow":    types.StringValue(sched["dow"].(string)),
			},
		)
		if !d.HasError() {
			model.Schedule = scheduleObj
		}
	}

	return nil
}
