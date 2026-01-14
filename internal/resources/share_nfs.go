package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/trueform/terraform-provider-trueform/internal/client"
)

var (
	_ resource.Resource                = &ShareNFSResource{}
	_ resource.ResourceWithImportState = &ShareNFSResource{}
)

func NewShareNFSResource() resource.Resource {
	return &ShareNFSResource{}
}

type ShareNFSResource struct {
	client *client.Client
}

type ShareNFSResourceModel struct {
	ID            types.Int64  `tfsdk:"id"`
	Path          types.String `tfsdk:"path"`
	Aliases       types.List   `tfsdk:"aliases"`
	Comment       types.String `tfsdk:"comment"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Networks      types.List   `tfsdk:"networks"`
	Hosts         types.List   `tfsdk:"hosts"`
	MaprootUser   types.String `tfsdk:"maproot_user"`
	MaprootGroup  types.String `tfsdk:"maproot_group"`
	MapallUser    types.String `tfsdk:"mapall_user"`
	MapallGroup   types.String `tfsdk:"mapall_group"`
	Security      types.List   `tfsdk:"security"`
	Ro            types.Bool   `tfsdk:"ro"`
	Locked        types.Bool   `tfsdk:"locked"`
}

func (r *ShareNFSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_share_nfs"
}

func (r *ShareNFSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an NFS share on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier for the share.",
				Computed:    true,
			},
			"path": schema.StringAttribute{
				Description: "The path to share.",
				Required:    true,
			},
			"aliases": schema.ListAttribute{
				Description: "List of aliases for the share.",
				Optional:    true,
				ElementType: types.StringType,
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
			"networks": schema.ListAttribute{
				Description: "List of authorized networks (CIDR notation).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"hosts": schema.ListAttribute{
				Description: "List of authorized hosts.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"maproot_user": schema.StringAttribute{
				Description: "Map root user to this user.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"maproot_group": schema.StringAttribute{
				Description: "Map root group to this group.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"mapall_user": schema.StringAttribute{
				Description: "Map all users to this user.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"mapall_group": schema.StringAttribute{
				Description: "Map all groups to this group.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"security": schema.ListAttribute{
				Description: "Security flavors (sys, krb5, krb5i, krb5p).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"ro": schema.BoolAttribute{
				Description: "Export share as read-only.",
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

func (r *ShareNFSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ShareNFSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ShareNFSResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating NFS share", map[string]interface{}{
		"path": plan.Path.ValueString(),
	})

	createData := map[string]interface{}{
		"path":    plan.Path.ValueString(),
		"enabled": plan.Enabled.ValueBool(),
	}

	if !plan.Aliases.IsNull() {
		var aliases []string
		diags = plan.Aliases.ElementsAs(ctx, &aliases, false)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			createData["aliases"] = aliases
		}
	}
	if !plan.Comment.IsNull() {
		createData["comment"] = plan.Comment.ValueString()
	}
	if !plan.Networks.IsNull() {
		var networks []string
		diags = plan.Networks.ElementsAs(ctx, &networks, false)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			createData["networks"] = networks
		}
	}
	if !plan.Hosts.IsNull() {
		var hosts []string
		diags = plan.Hosts.ElementsAs(ctx, &hosts, false)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			createData["hosts"] = hosts
		}
	}
	if !plan.MaprootUser.IsNull() && plan.MaprootUser.ValueString() != "" {
		createData["maproot_user"] = plan.MaprootUser.ValueString()
	}
	if !plan.MaprootGroup.IsNull() && plan.MaprootGroup.ValueString() != "" {
		createData["maproot_group"] = plan.MaprootGroup.ValueString()
	}
	if !plan.MapallUser.IsNull() && plan.MapallUser.ValueString() != "" {
		createData["mapall_user"] = plan.MapallUser.ValueString()
	}
	if !plan.MapallGroup.IsNull() && plan.MapallGroup.ValueString() != "" {
		createData["mapall_group"] = plan.MapallGroup.ValueString()
	}
	if !plan.Security.IsNull() {
		var security []string
		diags = plan.Security.ElementsAs(ctx, &security, false)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			createData["security"] = security
		}
	}
	if !plan.Ro.IsNull() {
		createData["ro"] = plan.Ro.ValueBool()
	}

	var result map[string]interface{}
	err := r.client.Create(ctx, "sharing.nfs", createData, &result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating NFS Share",
			"Could not create NFS share: "+err.Error(),
		)
		return
	}

	shareID := int64(result["id"].(float64))
	if err := r.readShare(ctx, shareID, &plan); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading NFS Share",
			"Could not read NFS share after creation: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ShareNFSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ShareNFSResourceModel
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
			"Error Reading NFS Share",
			"Could not read NFS share: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ShareNFSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ShareNFSResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ShareNFSResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating NFS share", map[string]interface{}{
		"id": state.ID.ValueInt64(),
	})

	updateData := map[string]interface{}{}

	if !plan.Path.Equal(state.Path) {
		updateData["path"] = plan.Path.ValueString()
	}
	if !plan.Aliases.Equal(state.Aliases) {
		var aliases []string
		if !plan.Aliases.IsNull() {
			diags = plan.Aliases.ElementsAs(ctx, &aliases, false)
			resp.Diagnostics.Append(diags...)
		}
		updateData["aliases"] = aliases
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
	if !plan.Networks.Equal(state.Networks) {
		var networks []string
		if !plan.Networks.IsNull() {
			diags = plan.Networks.ElementsAs(ctx, &networks, false)
			resp.Diagnostics.Append(diags...)
		}
		updateData["networks"] = networks
	}
	if !plan.Hosts.Equal(state.Hosts) {
		var hosts []string
		if !plan.Hosts.IsNull() {
			diags = plan.Hosts.ElementsAs(ctx, &hosts, false)
			resp.Diagnostics.Append(diags...)
		}
		updateData["hosts"] = hosts
	}
	if !plan.MaprootUser.Equal(state.MaprootUser) {
		updateData["maproot_user"] = plan.MaprootUser.ValueString()
	}
	if !plan.MaprootGroup.Equal(state.MaprootGroup) {
		updateData["maproot_group"] = plan.MaprootGroup.ValueString()
	}
	if !plan.MapallUser.Equal(state.MapallUser) {
		updateData["mapall_user"] = plan.MapallUser.ValueString()
	}
	if !plan.MapallGroup.Equal(state.MapallGroup) {
		updateData["mapall_group"] = plan.MapallGroup.ValueString()
	}
	if !plan.Security.Equal(state.Security) {
		var security []string
		if !plan.Security.IsNull() {
			diags = plan.Security.ElementsAs(ctx, &security, false)
			resp.Diagnostics.Append(diags...)
		}
		updateData["security"] = security
	}
	if !plan.Ro.Equal(state.Ro) {
		updateData["ro"] = plan.Ro.ValueBool()
	}

	if len(updateData) > 0 {
		var result map[string]interface{}
		err := r.client.Update(ctx, "sharing.nfs", state.ID.ValueInt64(), updateData, &result)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating NFS Share",
				"Could not update NFS share: "+err.Error(),
			)
			return
		}
	}

	if err := r.readShare(ctx, state.ID.ValueInt64(), &plan); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading NFS Share",
			"Could not read NFS share after update: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ShareNFSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ShareNFSResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting NFS share", map[string]interface{}{
		"id": state.ID.ValueInt64(),
	})

	err := r.client.Delete(ctx, "sharing.nfs", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting NFS Share",
			"Could not delete NFS share: "+err.Error(),
		)
		return
	}
}

func (r *ShareNFSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *ShareNFSResource) readShare(ctx context.Context, id int64, model *ShareNFSResourceModel) error {
	var result map[string]interface{}
	err := r.client.GetInstance(ctx, "sharing.nfs", id, &result)
	if err != nil {
		return err
	}

	model.ID = types.Int64Value(int64(result["id"].(float64)))
	model.Path = types.StringValue(result["path"].(string))

	if aliases, ok := result["aliases"].([]interface{}); ok {
		aliasList := make([]string, len(aliases))
		for i, a := range aliases {
			aliasList[i] = a.(string)
		}
		aliasValues, diags := types.ListValueFrom(ctx, types.StringType, aliasList)
		if !diags.HasError() {
			model.Aliases = aliasValues
		}
	}
	if comment, ok := result["comment"].(string); ok {
		model.Comment = types.StringValue(comment)
	}
	if enabled, ok := result["enabled"].(bool); ok {
		model.Enabled = types.BoolValue(enabled)
	}
	if networks, ok := result["networks"].([]interface{}); ok {
		networkList := make([]string, len(networks))
		for i, n := range networks {
			networkList[i] = n.(string)
		}
		networkValues, diags := types.ListValueFrom(ctx, types.StringType, networkList)
		if !diags.HasError() {
			model.Networks = networkValues
		}
	}
	if hosts, ok := result["hosts"].([]interface{}); ok {
		hostList := make([]string, len(hosts))
		for i, h := range hosts {
			hostList[i] = h.(string)
		}
		hostValues, diags := types.ListValueFrom(ctx, types.StringType, hostList)
		if !diags.HasError() {
			model.Hosts = hostValues
		}
	}
	if maprootUser, ok := result["maproot_user"].(string); ok {
		model.MaprootUser = types.StringValue(maprootUser)
	}
	if maprootGroup, ok := result["maproot_group"].(string); ok {
		model.MaprootGroup = types.StringValue(maprootGroup)
	}
	if mapallUser, ok := result["mapall_user"].(string); ok {
		model.MapallUser = types.StringValue(mapallUser)
	}
	if mapallGroup, ok := result["mapall_group"].(string); ok {
		model.MapallGroup = types.StringValue(mapallGroup)
	}
	if security, ok := result["security"].([]interface{}); ok {
		secList := make([]string, len(security))
		for i, s := range security {
			secList[i] = s.(string)
		}
		secValues, diags := types.ListValueFrom(ctx, types.StringType, secList)
		if !diags.HasError() {
			model.Security = secValues
		}
	}
	if ro, ok := result["ro"].(bool); ok {
		model.Ro = types.BoolValue(ro)
	}
	if locked, ok := result["locked"].(bool); ok {
		model.Locked = types.BoolValue(locked)
	}

	return nil
}
