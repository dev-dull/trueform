package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/trueform/terraform-provider-trueform/internal/client"
)

var (
	_ resource.Resource                = &ShareSMBResource{}
	_ resource.ResourceWithImportState = &ShareSMBResource{}
)

func NewShareSMBResource() resource.Resource {
	return &ShareSMBResource{}
}

type ShareSMBResource struct {
	client *client.Client
}

type ShareSMBResourceModel struct {
	ID                types.Int64  `tfsdk:"id"`
	Path              types.String `tfsdk:"path"`
	PathSuffix        types.String `tfsdk:"path_suffix"`
	Name              types.String `tfsdk:"name"`
	Comment           types.String `tfsdk:"comment"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	Home              types.Bool   `tfsdk:"home"`
	Purpose           types.String `tfsdk:"purpose"`
	TimeMachine       types.Bool   `tfsdk:"timemachine"`
	Ro                types.Bool   `tfsdk:"ro"`
	Browsable         types.Bool   `tfsdk:"browsable"`
	Recyclebin        types.Bool   `tfsdk:"recyclebin"`
	Guestok           types.Bool   `tfsdk:"guestok"`
	Abe               types.Bool   `tfsdk:"abe"`
	HostsAllow        types.List   `tfsdk:"hostsallow"`
	HostsDeny         types.List   `tfsdk:"hostsdeny"`
	AuxSMBConf        types.String `tfsdk:"auxsmbconf"`
	Acl               types.Bool   `tfsdk:"acl"`
	Durablehandle     types.Bool   `tfsdk:"durablehandle"`
	Shadowcopy        types.Bool   `tfsdk:"shadowcopy"`
	Streams           types.Bool   `tfsdk:"streams"`
	Fsrvp             types.Bool   `tfsdk:"fsrvp"`
	AuditLogging      types.Bool   `tfsdk:"audit_logging"`
	Locked            types.Bool   `tfsdk:"locked"`
}

func (r *ShareSMBResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_share_smb"
}

func (r *ShareSMBResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an SMB share on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier for the share.",
				Computed:    true,
			},
			"path": schema.StringAttribute{
				Description: "The path to share.",
				Required:    true,
			},
			"path_suffix": schema.StringAttribute{
				Description: "Suffix to append to the path.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the share.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"comment": schema.StringAttribute{
				Description: "Comment for the share.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the share is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"home": schema.BoolAttribute{
				Description: "Whether this is a home share.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"purpose": schema.StringAttribute{
				Description: "Purpose preset for the share.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("NO_PRESET"),
			},
			"timemachine": schema.BoolAttribute{
				Description: "Enable Time Machine support.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"ro": schema.BoolAttribute{
				Description: "Export share as read-only.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"browsable": schema.BoolAttribute{
				Description: "Whether the share is browsable.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"recyclebin": schema.BoolAttribute{
				Description: "Enable recycle bin.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"guestok": schema.BoolAttribute{
				Description: "Allow guest access.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"abe": schema.BoolAttribute{
				Description: "Enable Access Based Enumeration.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"hostsallow": schema.ListAttribute{
				Description: "List of allowed hosts.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"hostsdeny": schema.ListAttribute{
				Description: "List of denied hosts.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"auxsmbconf": schema.StringAttribute{
				Description: "Auxiliary SMB configuration parameters.",
				Optional:    true,
			},
			"acl": schema.BoolAttribute{
				Description: "Enable ACL support.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"durablehandle": schema.BoolAttribute{
				Description: "Enable durable handles.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"shadowcopy": schema.BoolAttribute{
				Description: "Enable shadow copies.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"streams": schema.BoolAttribute{
				Description: "Enable NTFS streams.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"fsrvp": schema.BoolAttribute{
				Description: "Enable File Server Remote VSS Protocol.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"audit_logging": schema.BoolAttribute{
				Description: "Enable audit logging.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"locked": schema.BoolAttribute{
				Description: "Whether the share is locked.",
				Computed:    true,
			},
		},
	}
}

func (r *ShareSMBResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ShareSMBResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ShareSMBResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating SMB share", map[string]interface{}{
		"name": plan.Name.ValueString(),
		"path": plan.Path.ValueString(),
	})

	createData := map[string]interface{}{
		"path":    plan.Path.ValueString(),
		"name":    plan.Name.ValueString(),
		"enabled": plan.Enabled.ValueBool(),
	}

	if !plan.PathSuffix.IsNull() {
		createData["path_suffix"] = plan.PathSuffix.ValueString()
	}
	if !plan.Comment.IsNull() {
		createData["comment"] = plan.Comment.ValueString()
	}
	if !plan.Home.IsNull() {
		createData["home"] = plan.Home.ValueBool()
	}
	if !plan.Purpose.IsNull() {
		createData["purpose"] = plan.Purpose.ValueString()
	}
	if !plan.TimeMachine.IsNull() {
		createData["timemachine"] = plan.TimeMachine.ValueBool()
	}
	if !plan.Ro.IsNull() {
		createData["ro"] = plan.Ro.ValueBool()
	}
	if !plan.Browsable.IsNull() {
		createData["browsable"] = plan.Browsable.ValueBool()
	}
	if !plan.Recyclebin.IsNull() {
		createData["recyclebin"] = plan.Recyclebin.ValueBool()
	}
	if !plan.Guestok.IsNull() {
		createData["guestok"] = plan.Guestok.ValueBool()
	}
	if !plan.Abe.IsNull() {
		createData["abe"] = plan.Abe.ValueBool()
	}
	if !plan.HostsAllow.IsNull() {
		var hosts []string
		diags = plan.HostsAllow.ElementsAs(ctx, &hosts, false)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			createData["hostsallow"] = hosts
		}
	}
	if !plan.HostsDeny.IsNull() {
		var hosts []string
		diags = plan.HostsDeny.ElementsAs(ctx, &hosts, false)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			createData["hostsdeny"] = hosts
		}
	}
	if !plan.AuxSMBConf.IsNull() {
		createData["auxsmbconf"] = plan.AuxSMBConf.ValueString()
	}
	if !plan.Acl.IsNull() {
		createData["acl"] = plan.Acl.ValueBool()
	}
	if !plan.Durablehandle.IsNull() {
		createData["durablehandle"] = plan.Durablehandle.ValueBool()
	}
	if !plan.Shadowcopy.IsNull() {
		createData["shadowcopy"] = plan.Shadowcopy.ValueBool()
	}
	if !plan.Streams.IsNull() {
		createData["streams"] = plan.Streams.ValueBool()
	}
	if !plan.Fsrvp.IsNull() {
		createData["fsrvp"] = plan.Fsrvp.ValueBool()
	}
	if !plan.AuditLogging.IsNull() {
		createData["audit_logging"] = plan.AuditLogging.ValueBool()
	}

	var result map[string]interface{}
	err := r.client.Create(ctx, "sharing.smb", createData, &result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating SMB Share",
			"Could not create SMB share: "+err.Error(),
		)
		return
	}

	shareID := int64(result["id"].(float64))
	if err := r.readShare(ctx, shareID, &plan); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SMB Share",
			"Could not read SMB share after creation: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ShareSMBResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ShareSMBResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readShare(ctx, state.ID.ValueInt64(), &state); err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading SMB Share",
			"Could not read SMB share: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ShareSMBResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ShareSMBResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ShareSMBResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating SMB share", map[string]interface{}{
		"id": state.ID.ValueInt64(),
	})

	updateData := map[string]interface{}{}

	if !plan.Path.Equal(state.Path) {
		updateData["path"] = plan.Path.ValueString()
	}
	if !plan.PathSuffix.Equal(state.PathSuffix) {
		if plan.PathSuffix.IsNull() {
			updateData["path_suffix"] = ""
		} else {
			updateData["path_suffix"] = plan.PathSuffix.ValueString()
		}
	}
	if !plan.Comment.Equal(state.Comment) {
		if plan.Comment.IsNull() {
			updateData["comment"] = ""
		} else {
			updateData["comment"] = plan.Comment.ValueString()
		}
	}
	if !plan.Enabled.Equal(state.Enabled) {
		updateData["enabled"] = plan.Enabled.ValueBool()
	}
	if !plan.Home.Equal(state.Home) {
		updateData["home"] = plan.Home.ValueBool()
	}
	if !plan.Purpose.Equal(state.Purpose) {
		updateData["purpose"] = plan.Purpose.ValueString()
	}
	if !plan.TimeMachine.Equal(state.TimeMachine) {
		updateData["timemachine"] = plan.TimeMachine.ValueBool()
	}
	if !plan.Ro.Equal(state.Ro) {
		updateData["ro"] = plan.Ro.ValueBool()
	}
	if !plan.Browsable.Equal(state.Browsable) {
		updateData["browsable"] = plan.Browsable.ValueBool()
	}
	if !plan.Recyclebin.Equal(state.Recyclebin) {
		updateData["recyclebin"] = plan.Recyclebin.ValueBool()
	}
	if !plan.Guestok.Equal(state.Guestok) {
		updateData["guestok"] = plan.Guestok.ValueBool()
	}
	if !plan.Abe.Equal(state.Abe) {
		updateData["abe"] = plan.Abe.ValueBool()
	}
	if !plan.HostsAllow.Equal(state.HostsAllow) {
		var hosts []string
		if !plan.HostsAllow.IsNull() {
			diags = plan.HostsAllow.ElementsAs(ctx, &hosts, false)
			resp.Diagnostics.Append(diags...)
		}
		updateData["hostsallow"] = hosts
	}
	if !plan.HostsDeny.Equal(state.HostsDeny) {
		var hosts []string
		if !plan.HostsDeny.IsNull() {
			diags = plan.HostsDeny.ElementsAs(ctx, &hosts, false)
			resp.Diagnostics.Append(diags...)
		}
		updateData["hostsdeny"] = hosts
	}
	if !plan.AuxSMBConf.Equal(state.AuxSMBConf) {
		if plan.AuxSMBConf.IsNull() {
			updateData["auxsmbconf"] = ""
		} else {
			updateData["auxsmbconf"] = plan.AuxSMBConf.ValueString()
		}
	}
	if !plan.Acl.Equal(state.Acl) {
		updateData["acl"] = plan.Acl.ValueBool()
	}
	if !plan.Durablehandle.Equal(state.Durablehandle) {
		updateData["durablehandle"] = plan.Durablehandle.ValueBool()
	}
	if !plan.Shadowcopy.Equal(state.Shadowcopy) {
		updateData["shadowcopy"] = plan.Shadowcopy.ValueBool()
	}
	if !plan.Streams.Equal(state.Streams) {
		updateData["streams"] = plan.Streams.ValueBool()
	}
	if !plan.Fsrvp.Equal(state.Fsrvp) {
		updateData["fsrvp"] = plan.Fsrvp.ValueBool()
	}
	if !plan.AuditLogging.Equal(state.AuditLogging) {
		updateData["audit_logging"] = plan.AuditLogging.ValueBool()
	}

	if len(updateData) > 0 {
		var result map[string]interface{}
		err := r.client.Update(ctx, "sharing.smb", state.ID.ValueInt64(), updateData, &result)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating SMB Share",
				"Could not update SMB share: "+err.Error(),
			)
			return
		}
	}

	if err := r.readShare(ctx, state.ID.ValueInt64(), &plan); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading SMB Share",
			"Could not read SMB share after update: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ShareSMBResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ShareSMBResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting SMB share", map[string]interface{}{
		"id": state.ID.ValueInt64(),
	})

	err := r.client.Delete(ctx, "sharing.smb", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting SMB Share",
			"Could not delete SMB share: "+err.Error(),
		)
		return
	}
}

func (r *ShareSMBResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *ShareSMBResource) readShare(ctx context.Context, id int64, model *ShareSMBResourceModel) error {
	var result map[string]interface{}
	err := r.client.GetInstance(ctx, "sharing.smb", id, &result)
	if err != nil {
		return err
	}

	model.ID = types.Int64Value(int64(result["id"].(float64)))
	model.Path = types.StringValue(result["path"].(string))
	model.Name = types.StringValue(result["name"].(string))

	if pathSuffix, ok := result["path_suffix"].(string); ok && pathSuffix != "" {
		model.PathSuffix = types.StringValue(pathSuffix)
	}
	if comment, ok := result["comment"].(string); ok {
		model.Comment = types.StringValue(comment)
	}
	if enabled, ok := result["enabled"].(bool); ok {
		model.Enabled = types.BoolValue(enabled)
	}
	if home, ok := result["home"].(bool); ok {
		model.Home = types.BoolValue(home)
	}
	if purpose, ok := result["purpose"].(string); ok {
		model.Purpose = types.StringValue(purpose)
	}
	if timemachine, ok := result["timemachine"].(bool); ok {
		model.TimeMachine = types.BoolValue(timemachine)
	}
	if ro, ok := result["ro"].(bool); ok {
		model.Ro = types.BoolValue(ro)
	}
	if browsable, ok := result["browsable"].(bool); ok {
		model.Browsable = types.BoolValue(browsable)
	}
	if recyclebin, ok := result["recyclebin"].(bool); ok {
		model.Recyclebin = types.BoolValue(recyclebin)
	}
	if guestok, ok := result["guestok"].(bool); ok {
		model.Guestok = types.BoolValue(guestok)
	}
	if abe, ok := result["abe"].(bool); ok {
		model.Abe = types.BoolValue(abe)
	}
	if hostsallow, ok := result["hostsallow"].([]interface{}); ok {
		hosts := make([]string, len(hostsallow))
		for i, h := range hostsallow {
			hosts[i] = h.(string)
		}
		hostValues, diags := types.ListValueFrom(ctx, types.StringType, hosts)
		if !diags.HasError() {
			model.HostsAllow = hostValues
		}
	}
	if hostsdeny, ok := result["hostsdeny"].([]interface{}); ok {
		hosts := make([]string, len(hostsdeny))
		for i, h := range hostsdeny {
			hosts[i] = h.(string)
		}
		hostValues, diags := types.ListValueFrom(ctx, types.StringType, hosts)
		if !diags.HasError() {
			model.HostsDeny = hostValues
		}
	}
	if auxsmbconf, ok := result["auxsmbconf"].(string); ok {
		model.AuxSMBConf = types.StringValue(auxsmbconf)
	}
	if acl, ok := result["acl"].(bool); ok {
		model.Acl = types.BoolValue(acl)
	}
	if durablehandle, ok := result["durablehandle"].(bool); ok {
		model.Durablehandle = types.BoolValue(durablehandle)
	}
	if shadowcopy, ok := result["shadowcopy"].(bool); ok {
		model.Shadowcopy = types.BoolValue(shadowcopy)
	}
	if streams, ok := result["streams"].(bool); ok {
		model.Streams = types.BoolValue(streams)
	}
	if fsrvp, ok := result["fsrvp"].(bool); ok {
		model.Fsrvp = types.BoolValue(fsrvp)
	}
	if auditLogging, ok := result["audit_logging"].(bool); ok {
		model.AuditLogging = types.BoolValue(auditLogging)
	}
	if locked, ok := result["locked"].(bool); ok {
		model.Locked = types.BoolValue(locked)
	}

	return nil
}
