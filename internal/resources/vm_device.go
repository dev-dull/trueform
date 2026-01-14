package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/trueform/terraform-provider-trueform/internal/client"
)

var (
	_ resource.Resource                = &VMDeviceResource{}
	_ resource.ResourceWithImportState = &VMDeviceResource{}
)

func NewVMDeviceResource() resource.Resource {
	return &VMDeviceResource{}
}

type VMDeviceResource struct {
	client *client.Client
}

type VMDeviceResourceModel struct {
	ID         types.Int64  `tfsdk:"id"`
	VM         types.Int64  `tfsdk:"vm"`
	DeviceType types.String `tfsdk:"dtype"`
	Order      types.Int64  `tfsdk:"order"`
	// Disk attributes
	DiskPath    types.String `tfsdk:"disk_path"`
	DiskType    types.String `tfsdk:"disk_type"`
	DiskSectorSize types.Int64 `tfsdk:"disk_sector_size"`
	// NIC attributes
	NICType     types.String `tfsdk:"nic_type"`
	NICMac      types.String `tfsdk:"nic_mac"`
	NICAttach   types.String `tfsdk:"nic_attach"`
	TrustGuestRXFilters types.Bool `tfsdk:"trust_guest_rx_filters"`
	// CDROM attributes
	CDROMPath   types.String `tfsdk:"cdrom_path"`
	// Display attributes
	DisplayType    types.String `tfsdk:"display_type"`
	DisplayPort    types.Int64  `tfsdk:"display_port"`
	DisplayBind    types.String `tfsdk:"display_bind"`
	DisplayPassword types.String `tfsdk:"display_password"`
	DisplayWeb     types.Bool   `tfsdk:"display_web"`
	DisplayResolution types.String `tfsdk:"display_resolution"`
	// PCI attributes
	PCIDevice   types.String `tfsdk:"pci_device"`
	// USB attributes
	USBDevice   types.String `tfsdk:"usb_device"`
	// RAW attributes
	RawSize     types.Int64  `tfsdk:"raw_size"`
	RawPath     types.String `tfsdk:"raw_path"`
}

func (r *VMDeviceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_device"
}

func (r *VMDeviceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a virtual machine device on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier for the device.",
				Computed:    true,
			},
			"vm": schema.Int64Attribute{
				Description: "The ID of the VM this device belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					// int64planmodifier.RequiresReplace(),
				},
			},
			"dtype": schema.StringAttribute{
				Description: "Device type (DISK, NIC, CDROM, DISPLAY, PCI, USB, RAW).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"order": schema.Int64Attribute{
				Description: "Boot order for the device.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1000),
			},
			// Disk attributes
			"disk_path": schema.StringAttribute{
				Description: "Path to the zvol for DISK type.",
				Optional:    true,
			},
			"disk_type": schema.StringAttribute{
				Description: "Disk type (AHCI, VIRTIO).",
				Optional:    true,
			},
			"disk_sector_size": schema.Int64Attribute{
				Description: "Disk sector size (512, 4096).",
				Optional:    true,
			},
			// NIC attributes
			"nic_type": schema.StringAttribute{
				Description: "NIC type (E1000, VIRTIO).",
				Optional:    true,
			},
			"nic_mac": schema.StringAttribute{
				Description: "MAC address for the NIC.",
				Optional:    true,
				Computed:    true,
			},
			"nic_attach": schema.StringAttribute{
				Description: "Network interface to attach to.",
				Optional:    true,
			},
			"trust_guest_rx_filters": schema.BoolAttribute{
				Description: "Trust guest RX filters.",
				Optional:    true,
			},
			// CDROM attributes
			"cdrom_path": schema.StringAttribute{
				Description: "Path to ISO file for CDROM type.",
				Optional:    true,
			},
			// Display attributes
			"display_type": schema.StringAttribute{
				Description: "Display type (VNC, SPICE).",
				Optional:    true,
			},
			"display_port": schema.Int64Attribute{
				Description: "Display port number.",
				Optional:    true,
			},
			"display_bind": schema.StringAttribute{
				Description: "IP address to bind display to.",
				Optional:    true,
			},
			"display_password": schema.StringAttribute{
				Description: "Display password.",
				Optional:    true,
				Sensitive:   true,
			},
			"display_web": schema.BoolAttribute{
				Description: "Enable web interface for display.",
				Optional:    true,
			},
			"display_resolution": schema.StringAttribute{
				Description: "Display resolution.",
				Optional:    true,
			},
			// PCI attributes
			"pci_device": schema.StringAttribute{
				Description: "PCI device identifier for passthrough.",
				Optional:    true,
			},
			// USB attributes
			"usb_device": schema.StringAttribute{
				Description: "USB device identifier for passthrough.",
				Optional:    true,
			},
			// RAW attributes
			"raw_size": schema.Int64Attribute{
				Description: "Size for RAW device.",
				Optional:    true,
			},
			"raw_path": schema.StringAttribute{
				Description: "Path for RAW file device.",
				Optional:    true,
			},
		},
	}
}

func (r *VMDeviceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VMDeviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VMDeviceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating VM device", map[string]interface{}{
		"vm":    plan.VM.ValueInt64(),
		"dtype": plan.DeviceType.ValueString(),
	})

	createData := map[string]interface{}{
		"vm":    plan.VM.ValueInt64(),
		"dtype": plan.DeviceType.ValueString(),
		"order": plan.Order.ValueInt64(),
	}

	// Build attributes based on device type
	attrs := map[string]interface{}{}

	switch plan.DeviceType.ValueString() {
	case "DISK":
		if !plan.DiskPath.IsNull() {
			attrs["path"] = plan.DiskPath.ValueString()
		}
		if !plan.DiskType.IsNull() {
			attrs["type"] = plan.DiskType.ValueString()
		}
		if !plan.DiskSectorSize.IsNull() {
			attrs["physical_sectorsize"] = plan.DiskSectorSize.ValueInt64()
			attrs["logical_sectorsize"] = plan.DiskSectorSize.ValueInt64()
		}
	case "NIC":
		if !plan.NICType.IsNull() {
			attrs["type"] = plan.NICType.ValueString()
		}
		if !plan.NICMac.IsNull() {
			attrs["mac"] = plan.NICMac.ValueString()
		}
		if !plan.NICAttach.IsNull() {
			attrs["nic_attach"] = plan.NICAttach.ValueString()
		}
		if !plan.TrustGuestRXFilters.IsNull() {
			attrs["trust_guest_rx_filters"] = plan.TrustGuestRXFilters.ValueBool()
		}
	case "CDROM":
		if !plan.CDROMPath.IsNull() {
			attrs["path"] = plan.CDROMPath.ValueString()
		}
	case "DISPLAY":
		if !plan.DisplayType.IsNull() {
			attrs["type"] = plan.DisplayType.ValueString()
		}
		if !plan.DisplayPort.IsNull() {
			attrs["port"] = plan.DisplayPort.ValueInt64()
		}
		if !plan.DisplayBind.IsNull() {
			attrs["bind"] = plan.DisplayBind.ValueString()
		}
		if !plan.DisplayPassword.IsNull() {
			attrs["password"] = plan.DisplayPassword.ValueString()
		}
		if !plan.DisplayWeb.IsNull() {
			attrs["web"] = plan.DisplayWeb.ValueBool()
		}
		if !plan.DisplayResolution.IsNull() {
			attrs["resolution"] = plan.DisplayResolution.ValueString()
		}
	case "PCI":
		if !plan.PCIDevice.IsNull() {
			attrs["pptdev"] = plan.PCIDevice.ValueString()
		}
	case "USB":
		if !plan.USBDevice.IsNull() {
			attrs["device"] = plan.USBDevice.ValueString()
		}
	case "RAW":
		if !plan.RawSize.IsNull() {
			attrs["size"] = plan.RawSize.ValueInt64()
		}
		if !plan.RawPath.IsNull() {
			attrs["path"] = plan.RawPath.ValueString()
		}
	}

	createData["attributes"] = attrs

	var result map[string]interface{}
	err := r.client.Create(ctx, "vm.device", createData, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating VM Device", "Could not create VM device: "+err.Error())
		return
	}

	deviceID := int64(result["id"].(float64))
	if err := r.readDevice(ctx, deviceID, &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading VM Device", "Could not read VM device after creation: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *VMDeviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VMDeviceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readDevice(ctx, state.ID.ValueInt64(), &state); err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading VM Device", "Could not read VM device: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *VMDeviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VMDeviceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state VMDeviceResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateData := map[string]interface{}{
		"order": plan.Order.ValueInt64(),
	}

	attrs := map[string]interface{}{}

	switch plan.DeviceType.ValueString() {
	case "DISK":
		if !plan.DiskPath.IsNull() {
			attrs["path"] = plan.DiskPath.ValueString()
		}
		if !plan.DiskType.IsNull() {
			attrs["type"] = plan.DiskType.ValueString()
		}
	case "NIC":
		if !plan.NICType.IsNull() {
			attrs["type"] = plan.NICType.ValueString()
		}
		if !plan.NICAttach.IsNull() {
			attrs["nic_attach"] = plan.NICAttach.ValueString()
		}
	case "CDROM":
		if !plan.CDROMPath.IsNull() {
			attrs["path"] = plan.CDROMPath.ValueString()
		}
	case "DISPLAY":
		if !plan.DisplayPassword.IsNull() {
			attrs["password"] = plan.DisplayPassword.ValueString()
		}
		if !plan.DisplayWeb.IsNull() {
			attrs["web"] = plan.DisplayWeb.ValueBool()
		}
	}

	if len(attrs) > 0 {
		updateData["attributes"] = attrs
	}

	var result map[string]interface{}
	err := r.client.Update(ctx, "vm.device", state.ID.ValueInt64(), updateData, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating VM Device", "Could not update VM device: "+err.Error())
		return
	}

	if err := r.readDevice(ctx, state.ID.ValueInt64(), &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading VM Device", "Could not read VM device after update: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *VMDeviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VMDeviceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "vm.device", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting VM Device", "Could not delete VM device: "+err.Error())
		return
	}
}

func (r *VMDeviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *VMDeviceResource) readDevice(ctx context.Context, id int64, model *VMDeviceResourceModel) error {
	var result map[string]interface{}
	err := r.client.GetInstance(ctx, "vm.device", id, &result)
	if err != nil {
		return err
	}

	model.ID = types.Int64Value(int64(result["id"].(float64)))
	model.VM = types.Int64Value(int64(result["vm"].(float64)))
	model.DeviceType = types.StringValue(result["dtype"].(string))

	if order, ok := result["order"].(float64); ok {
		model.Order = types.Int64Value(int64(order))
	}

	if attrs, ok := result["attributes"].(map[string]interface{}); ok {
		switch model.DeviceType.ValueString() {
		case "DISK":
			if path, ok := attrs["path"].(string); ok {
				model.DiskPath = types.StringValue(path)
			}
			if dtype, ok := attrs["type"].(string); ok {
				model.DiskType = types.StringValue(dtype)
			}
		case "NIC":
			if nicType, ok := attrs["type"].(string); ok {
				model.NICType = types.StringValue(nicType)
			}
			if mac, ok := attrs["mac"].(string); ok {
				model.NICMac = types.StringValue(mac)
			}
			if attach, ok := attrs["nic_attach"].(string); ok {
				model.NICAttach = types.StringValue(attach)
			}
		case "CDROM":
			if path, ok := attrs["path"].(string); ok {
				model.CDROMPath = types.StringValue(path)
			}
		case "DISPLAY":
			if displayType, ok := attrs["type"].(string); ok {
				model.DisplayType = types.StringValue(displayType)
			}
			if port, ok := attrs["port"].(float64); ok {
				model.DisplayPort = types.Int64Value(int64(port))
			}
			if bind, ok := attrs["bind"].(string); ok {
				model.DisplayBind = types.StringValue(bind)
			}
			if web, ok := attrs["web"].(bool); ok {
				model.DisplayWeb = types.BoolValue(web)
			}
			if resolution, ok := attrs["resolution"].(string); ok {
				model.DisplayResolution = types.StringValue(resolution)
			}
		case "PCI":
			if pptdev, ok := attrs["pptdev"].(string); ok {
				model.PCIDevice = types.StringValue(pptdev)
			}
		}
	}

	return nil
}
