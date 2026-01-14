package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/trueform/terraform-provider-trueform/internal/client"
)

var (
	_ resource.Resource                = &ISCSIExtentResource{}
	_ resource.ResourceWithImportState = &ISCSIExtentResource{}
)

func NewISCSIExtentResource() resource.Resource {
	return &ISCSIExtentResource{}
}

type ISCSIExtentResource struct {
	client *client.Client
}

type ISCSIExtentResourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	Disk         types.String `tfsdk:"disk"`
	Path         types.String `tfsdk:"path"`
	Filesize     types.Int64  `tfsdk:"filesize"`
	Blocksize    types.Int64  `tfsdk:"blocksize"`
	Pblocksize   types.Bool   `tfsdk:"pblocksize"`
	AvailThreshold types.Int64 `tfsdk:"avail_threshold"`
	Comment      types.String `tfsdk:"comment"`
	InsecureTPC  types.Bool   `tfsdk:"insecure_tpc"`
	Xen          types.Bool   `tfsdk:"xen"`
	RPM          types.String `tfsdk:"rpm"`
	Ro           types.Bool   `tfsdk:"ro"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	Serial       types.String `tfsdk:"serial"`
	NAA          types.String `tfsdk:"naa"`
	Locked       types.Bool   `tfsdk:"locked"`
}

func (r *ISCSIExtentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iscsi_extent"
}

func (r *ISCSIExtentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an iSCSI extent (LUN) on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier for the extent.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the extent.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Extent type (DISK or FILE).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"disk": schema.StringAttribute{
				Description: "Disk to use when type is DISK (zvol path).",
				Optional:    true,
			},
			"path": schema.StringAttribute{
				Description: "Path for file-based extent when type is FILE.",
				Optional:    true,
			},
			"filesize": schema.Int64Attribute{
				Description: "Size of the file extent in bytes.",
				Optional:    true,
			},
			"blocksize": schema.Int64Attribute{
				Description: "Logical block size (512 or 4096).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(512),
			},
			"pblocksize": schema.BoolAttribute{
				Description: "Use physical block size.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"avail_threshold": schema.Int64Attribute{
				Description: "Alert threshold for available space percentage.",
				Optional:    true,
			},
			"comment": schema.StringAttribute{
				Description: "Comment for the extent.",
				Optional:    true,
			},
			"insecure_tpc": schema.BoolAttribute{
				Description: "Allow Third Party Copy (TPC) operations.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"xen": schema.BoolAttribute{
				Description: "Enable Xen compatibility mode.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"rpm": schema.StringAttribute{
				Description: "Reported RPM (SSD, 5400, 7200, 10000, 15000).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("SSD"),
			},
			"ro": schema.BoolAttribute{
				Description: "Export extent as read-only.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the extent is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"serial": schema.StringAttribute{
				Description: "Serial number for the extent.",
				Computed:    true,
			},
			"naa": schema.StringAttribute{
				Description: "NAA identifier for the extent.",
				Computed:    true,
			},
			"locked": schema.BoolAttribute{
				Description: "Whether the extent is locked.",
				Computed:    true,
			},
		},
	}
}

func (r *ISCSIExtentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ISCSIExtentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ISCSIExtentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating iSCSI extent", map[string]interface{}{
		"name": plan.Name.ValueString(),
	})

	createData := map[string]interface{}{
		"name":    plan.Name.ValueString(),
		"type":    plan.Type.ValueString(),
		"enabled": plan.Enabled.ValueBool(),
	}

	if !plan.Disk.IsNull() {
		createData["disk"] = plan.Disk.ValueString()
	}
	if !plan.Path.IsNull() {
		createData["path"] = plan.Path.ValueString()
	}
	if !plan.Filesize.IsNull() {
		createData["filesize"] = plan.Filesize.ValueInt64()
	}
	if !plan.Blocksize.IsNull() {
		createData["blocksize"] = plan.Blocksize.ValueInt64()
	}
	if !plan.Pblocksize.IsNull() {
		createData["pblocksize"] = plan.Pblocksize.ValueBool()
	}
	if !plan.AvailThreshold.IsNull() {
		createData["avail_threshold"] = plan.AvailThreshold.ValueInt64()
	}
	if !plan.Comment.IsNull() {
		createData["comment"] = plan.Comment.ValueString()
	}
	if !plan.InsecureTPC.IsNull() {
		createData["insecure_tpc"] = plan.InsecureTPC.ValueBool()
	}
	if !plan.Xen.IsNull() {
		createData["xen"] = plan.Xen.ValueBool()
	}
	if !plan.RPM.IsNull() {
		createData["rpm"] = plan.RPM.ValueString()
	}
	if !plan.Ro.IsNull() {
		createData["ro"] = plan.Ro.ValueBool()
	}

	var result map[string]interface{}
	err := r.client.Create(ctx, "iscsi.extent", createData, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating iSCSI Extent", "Could not create iSCSI extent: "+err.Error())
		return
	}

	extentID := int64(result["id"].(float64))
	if err := r.readExtent(ctx, extentID, &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading iSCSI Extent", "Could not read iSCSI extent after creation: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ISCSIExtentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ISCSIExtentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readExtent(ctx, state.ID.ValueInt64(), &state); err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading iSCSI Extent", "Could not read iSCSI extent: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ISCSIExtentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ISCSIExtentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ISCSIExtentResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateData := map[string]interface{}{}

	if !plan.Disk.Equal(state.Disk) {
		updateData["disk"] = plan.Disk.ValueString()
	}
	if !plan.Path.Equal(state.Path) {
		updateData["path"] = plan.Path.ValueString()
	}
	if !plan.Filesize.Equal(state.Filesize) {
		updateData["filesize"] = plan.Filesize.ValueInt64()
	}
	if !plan.Blocksize.Equal(state.Blocksize) {
		updateData["blocksize"] = plan.Blocksize.ValueInt64()
	}
	if !plan.Pblocksize.Equal(state.Pblocksize) {
		updateData["pblocksize"] = plan.Pblocksize.ValueBool()
	}
	if !plan.AvailThreshold.Equal(state.AvailThreshold) {
		updateData["avail_threshold"] = plan.AvailThreshold.ValueInt64()
	}
	if !plan.Comment.Equal(state.Comment) {
		if plan.Comment.IsNull() {
			updateData["comment"] = ""
		} else {
			updateData["comment"] = plan.Comment.ValueString()
		}
	}
	if !plan.InsecureTPC.Equal(state.InsecureTPC) {
		updateData["insecure_tpc"] = plan.InsecureTPC.ValueBool()
	}
	if !plan.Xen.Equal(state.Xen) {
		updateData["xen"] = plan.Xen.ValueBool()
	}
	if !plan.RPM.Equal(state.RPM) {
		updateData["rpm"] = plan.RPM.ValueString()
	}
	if !plan.Ro.Equal(state.Ro) {
		updateData["ro"] = plan.Ro.ValueBool()
	}
	if !plan.Enabled.Equal(state.Enabled) {
		updateData["enabled"] = plan.Enabled.ValueBool()
	}

	if len(updateData) > 0 {
		var result map[string]interface{}
		err := r.client.Update(ctx, "iscsi.extent", state.ID.ValueInt64(), updateData, &result)
		if err != nil {
			resp.Diagnostics.AddError("Error Updating iSCSI Extent", "Could not update iSCSI extent: "+err.Error())
			return
		}
	}

	if err := r.readExtent(ctx, state.ID.ValueInt64(), &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading iSCSI Extent", "Could not read iSCSI extent after update: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ISCSIExtentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ISCSIExtentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "iscsi.extent", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting iSCSI Extent", "Could not delete iSCSI extent: "+err.Error())
		return
	}
}

func (r *ISCSIExtentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *ISCSIExtentResource) readExtent(ctx context.Context, id int64, model *ISCSIExtentResourceModel) error {
	var result map[string]interface{}
	err := r.client.GetInstance(ctx, "iscsi.extent", id, &result)
	if err != nil {
		return err
	}

	model.ID = types.Int64Value(int64(result["id"].(float64)))
	model.Name = types.StringValue(result["name"].(string))
	model.Type = types.StringValue(result["type"].(string))

	if disk, ok := result["disk"].(string); ok {
		model.Disk = types.StringValue(disk)
	}
	if path, ok := result["path"].(string); ok {
		model.Path = types.StringValue(path)
	}
	if filesize, ok := result["filesize"].(float64); ok {
		model.Filesize = types.Int64Value(int64(filesize))
	}
	if blocksize, ok := result["blocksize"].(float64); ok {
		model.Blocksize = types.Int64Value(int64(blocksize))
	}
	if pblocksize, ok := result["pblocksize"].(bool); ok {
		model.Pblocksize = types.BoolValue(pblocksize)
	}
	if availThreshold, ok := result["avail_threshold"].(float64); ok {
		model.AvailThreshold = types.Int64Value(int64(availThreshold))
	}
	if comment, ok := result["comment"].(string); ok {
		model.Comment = types.StringValue(comment)
	}
	if insecureTPC, ok := result["insecure_tpc"].(bool); ok {
		model.InsecureTPC = types.BoolValue(insecureTPC)
	}
	if xen, ok := result["xen"].(bool); ok {
		model.Xen = types.BoolValue(xen)
	}
	if rpm, ok := result["rpm"].(string); ok {
		model.RPM = types.StringValue(rpm)
	}
	if ro, ok := result["ro"].(bool); ok {
		model.Ro = types.BoolValue(ro)
	}
	if enabled, ok := result["enabled"].(bool); ok {
		model.Enabled = types.BoolValue(enabled)
	}
	if serial, ok := result["serial"].(string); ok {
		model.Serial = types.StringValue(serial)
	}
	if naa, ok := result["naa"].(string); ok {
		model.NAA = types.StringValue(naa)
	}
	if locked, ok := result["locked"].(bool); ok {
		model.Locked = types.BoolValue(locked)
	}

	return nil
}
