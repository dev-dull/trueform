package resources

import (
	"context"
	"fmt"

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
	_ resource.Resource                = &VMResource{}
	_ resource.ResourceWithImportState = &VMResource{}
)

func NewVMResource() resource.Resource {
	return &VMResource{}
}

type VMResource struct {
	client *client.Client
}

type VMResourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	VCPUs            types.Int64  `tfsdk:"vcpus"`
	Cores            types.Int64  `tfsdk:"cores"`
	Threads          types.Int64  `tfsdk:"threads"`
	Memory           types.Int64  `tfsdk:"memory"`
	MinMemory        types.Int64  `tfsdk:"min_memory"`
	Bootloader       types.String `tfsdk:"bootloader"`
	BootloaderOVMF   types.String `tfsdk:"bootloader_ovmf"`
	Autostart        types.Bool   `tfsdk:"autostart"`
	HideFromMSR      types.Bool   `tfsdk:"hide_from_msr"`
	EnsureDisplayDevice types.Bool `tfsdk:"ensure_display_device"`
	Time             types.String `tfsdk:"time"`
	ShutdownTimeout  types.Int64  `tfsdk:"shutdown_timeout"`
	ArchType         types.String `tfsdk:"arch_type"`
	MachineType      types.String `tfsdk:"machine_type"`
	UUID             types.String `tfsdk:"uuid"`
	CPUMode          types.String `tfsdk:"cpu_mode"`
	CPUModel         types.String `tfsdk:"cpu_model"`
	Status           types.String `tfsdk:"status"`
}

func (r *VMResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}

func (r *VMResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a virtual machine on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier for the VM.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the VM.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the VM.",
				Optional:    true,
			},
			"vcpus": schema.Int64Attribute{
				Description: "Number of virtual CPUs.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
			},
			"cores": schema.Int64Attribute{
				Description: "Number of cores per socket.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
			},
			"threads": schema.Int64Attribute{
				Description: "Number of threads per core.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
			},
			"memory": schema.Int64Attribute{
				Description: "Memory in MiB.",
				Required:    true,
			},
			"min_memory": schema.Int64Attribute{
				Description: "Minimum memory in MiB (for ballooning).",
				Optional:    true,
			},
			"bootloader": schema.StringAttribute{
				Description: "Bootloader type (UEFI, UEFI_CSM).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("UEFI"),
			},
			"bootloader_ovmf": schema.StringAttribute{
				Description: "OVMF firmware type (OVMF_CODE, OVMF_CODE_4M, etc.).",
				Optional:    true,
			},
			"autostart": schema.BoolAttribute{
				Description: "Whether to autostart the VM.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"hide_from_msr": schema.BoolAttribute{
				Description: "Hide VM from MSR.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"ensure_display_device": schema.BoolAttribute{
				Description: "Ensure a display device is present.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"time": schema.StringAttribute{
				Description: "Time setting (LOCAL, UTC).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("LOCAL"),
			},
			"shutdown_timeout": schema.Int64Attribute{
				Description: "Shutdown timeout in seconds.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(90),
			},
			"arch_type": schema.StringAttribute{
				Description: "CPU architecture type.",
				Optional:    true,
			},
			"machine_type": schema.StringAttribute{
				Description: "Machine type (pc, q35).",
				Optional:    true,
			},
			"uuid": schema.StringAttribute{
				Description: "VM UUID.",
				Computed:    true,
			},
			"cpu_mode": schema.StringAttribute{
				Description: "CPU mode (CUSTOM, HOST_MODEL, HOST_PASSTHROUGH).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("CUSTOM"),
			},
			"cpu_model": schema.StringAttribute{
				Description: "CPU model when cpu_mode is CUSTOM.",
				Optional:    true,
			},
			"status": schema.StringAttribute{
				Description: "Current status of the VM.",
				Computed:    true,
			},
		},
	}
}

func (r *VMResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VMResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating VM", map[string]interface{}{
		"name": plan.Name.ValueString(),
	})

	createData := map[string]interface{}{
		"name":       plan.Name.ValueString(),
		"vcpus":      plan.VCPUs.ValueInt64(),
		"cores":      plan.Cores.ValueInt64(),
		"threads":    plan.Threads.ValueInt64(),
		"memory":     plan.Memory.ValueInt64(),
		"bootloader": plan.Bootloader.ValueString(),
		"autostart":  plan.Autostart.ValueBool(),
	}

	if !plan.Description.IsNull() {
		createData["description"] = plan.Description.ValueString()
	}
	if !plan.MinMemory.IsNull() {
		createData["min_memory"] = plan.MinMemory.ValueInt64()
	}
	if !plan.BootloaderOVMF.IsNull() {
		createData["bootloader_ovmf"] = plan.BootloaderOVMF.ValueString()
	}
	if !plan.HideFromMSR.IsNull() {
		createData["hide_from_msr"] = plan.HideFromMSR.ValueBool()
	}
	if !plan.EnsureDisplayDevice.IsNull() {
		createData["ensure_display_device"] = plan.EnsureDisplayDevice.ValueBool()
	}
	if !plan.Time.IsNull() {
		createData["time"] = plan.Time.ValueString()
	}
	if !plan.ShutdownTimeout.IsNull() {
		createData["shutdown_timeout"] = plan.ShutdownTimeout.ValueInt64()
	}
	if !plan.ArchType.IsNull() {
		createData["arch_type"] = plan.ArchType.ValueString()
	}
	if !plan.MachineType.IsNull() {
		createData["machine_type"] = plan.MachineType.ValueString()
	}
	if !plan.CPUMode.IsNull() {
		createData["cpu_mode"] = plan.CPUMode.ValueString()
	}
	if !plan.CPUModel.IsNull() {
		createData["cpu_model"] = plan.CPUModel.ValueString()
	}

	var result map[string]interface{}
	err := r.client.Create(ctx, "vm", createData, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating VM", "Could not create VM: "+err.Error())
		return
	}

	vmID := int64(result["id"].(float64))
	if err := r.readVM(ctx, vmID, &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading VM", "Could not read VM after creation: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *VMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VMResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readVM(ctx, state.ID.ValueInt64(), &state); err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading VM", "Could not read VM: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *VMResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VMResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state VMResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateData := map[string]interface{}{}

	if !plan.Description.Equal(state.Description) {
		if plan.Description.IsNull() {
			updateData["description"] = ""
		} else {
			updateData["description"] = plan.Description.ValueString()
		}
	}
	if !plan.VCPUs.Equal(state.VCPUs) {
		updateData["vcpus"] = plan.VCPUs.ValueInt64()
	}
	if !plan.Cores.Equal(state.Cores) {
		updateData["cores"] = plan.Cores.ValueInt64()
	}
	if !plan.Threads.Equal(state.Threads) {
		updateData["threads"] = plan.Threads.ValueInt64()
	}
	if !plan.Memory.Equal(state.Memory) {
		updateData["memory"] = plan.Memory.ValueInt64()
	}
	if !plan.MinMemory.Equal(state.MinMemory) {
		updateData["min_memory"] = plan.MinMemory.ValueInt64()
	}
	if !plan.Bootloader.Equal(state.Bootloader) {
		updateData["bootloader"] = plan.Bootloader.ValueString()
	}
	if !plan.BootloaderOVMF.Equal(state.BootloaderOVMF) {
		updateData["bootloader_ovmf"] = plan.BootloaderOVMF.ValueString()
	}
	if !plan.Autostart.Equal(state.Autostart) {
		updateData["autostart"] = plan.Autostart.ValueBool()
	}
	if !plan.HideFromMSR.Equal(state.HideFromMSR) {
		updateData["hide_from_msr"] = plan.HideFromMSR.ValueBool()
	}
	if !plan.EnsureDisplayDevice.Equal(state.EnsureDisplayDevice) {
		updateData["ensure_display_device"] = plan.EnsureDisplayDevice.ValueBool()
	}
	if !plan.Time.Equal(state.Time) {
		updateData["time"] = plan.Time.ValueString()
	}
	if !plan.ShutdownTimeout.Equal(state.ShutdownTimeout) {
		updateData["shutdown_timeout"] = plan.ShutdownTimeout.ValueInt64()
	}
	if !plan.CPUMode.Equal(state.CPUMode) {
		updateData["cpu_mode"] = plan.CPUMode.ValueString()
	}
	if !plan.CPUModel.Equal(state.CPUModel) {
		updateData["cpu_model"] = plan.CPUModel.ValueString()
	}

	if len(updateData) > 0 {
		var result map[string]interface{}
		err := r.client.Update(ctx, "vm", state.ID.ValueInt64(), updateData, &result)
		if err != nil {
			resp.Diagnostics.AddError("Error Updating VM", "Could not update VM: "+err.Error())
			return
		}
	}

	if err := r.readVM(ctx, state.ID.ValueInt64(), &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading VM", "Could not read VM after update: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *VMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VMResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Stop the VM first if running
	r.client.Call(ctx, "vm.stop", []interface{}{state.ID.ValueInt64()}, nil)

	err := r.client.Delete(ctx, "vm", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting VM", "Could not delete VM: "+err.Error())
		return
	}
}

func (r *VMResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *VMResource) readVM(ctx context.Context, id int64, model *VMResourceModel) error {
	var result map[string]interface{}
	err := r.client.GetInstance(ctx, "vm", id, &result)
	if err != nil {
		return err
	}

	model.ID = types.Int64Value(int64(result["id"].(float64)))
	model.Name = types.StringValue(result["name"].(string))

	if description, ok := result["description"].(string); ok {
		model.Description = types.StringValue(description)
	}
	if vcpus, ok := result["vcpus"].(float64); ok {
		model.VCPUs = types.Int64Value(int64(vcpus))
	}
	if cores, ok := result["cores"].(float64); ok {
		model.Cores = types.Int64Value(int64(cores))
	}
	if threads, ok := result["threads"].(float64); ok {
		model.Threads = types.Int64Value(int64(threads))
	}
	if memory, ok := result["memory"].(float64); ok {
		model.Memory = types.Int64Value(int64(memory))
	}
	if minMemory, ok := result["min_memory"].(float64); ok {
		model.MinMemory = types.Int64Value(int64(minMemory))
	}
	if bootloader, ok := result["bootloader"].(string); ok {
		model.Bootloader = types.StringValue(bootloader)
	}
	if bootloaderOVMF, ok := result["bootloader_ovmf"].(string); ok {
		model.BootloaderOVMF = types.StringValue(bootloaderOVMF)
	}
	if autostart, ok := result["autostart"].(bool); ok {
		model.Autostart = types.BoolValue(autostart)
	}
	if hideFromMSR, ok := result["hide_from_msr"].(bool); ok {
		model.HideFromMSR = types.BoolValue(hideFromMSR)
	}
	if ensureDisplayDevice, ok := result["ensure_display_device"].(bool); ok {
		model.EnsureDisplayDevice = types.BoolValue(ensureDisplayDevice)
	}
	if time, ok := result["time"].(string); ok {
		model.Time = types.StringValue(time)
	}
	if shutdownTimeout, ok := result["shutdown_timeout"].(float64); ok {
		model.ShutdownTimeout = types.Int64Value(int64(shutdownTimeout))
	}
	if archType, ok := result["arch_type"].(string); ok {
		model.ArchType = types.StringValue(archType)
	}
	if machineType, ok := result["machine_type"].(string); ok {
		model.MachineType = types.StringValue(machineType)
	}
	if uuid, ok := result["uuid"].(string); ok {
		model.UUID = types.StringValue(uuid)
	}
	if cpuMode, ok := result["cpu_mode"].(string); ok {
		model.CPUMode = types.StringValue(cpuMode)
	}
	if cpuModel, ok := result["cpu_model"].(string); ok {
		model.CPUModel = types.StringValue(cpuModel)
	}
	if status, ok := result["status"].(map[string]interface{}); ok {
		if state, ok := status["state"].(string); ok {
			model.Status = types.StringValue(state)
		}
	}

	return nil
}
