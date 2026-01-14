package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/trueform/terraform-provider-trueform/internal/client"
)

var _ datasource.DataSource = &VMDataSource{}

func NewVMDataSource() datasource.DataSource {
	return &VMDataSource{}
}

type VMDataSource struct {
	client *client.Client
}

type VMDataSourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	VCPUs       types.Int64  `tfsdk:"vcpus"`
	Cores       types.Int64  `tfsdk:"cores"`
	Threads     types.Int64  `tfsdk:"threads"`
	Memory      types.Int64  `tfsdk:"memory"`
	Bootloader  types.String `tfsdk:"bootloader"`
	Autostart   types.Bool   `tfsdk:"autostart"`
	Status      types.String `tfsdk:"status"`
	UUID        types.String `tfsdk:"uuid"`
}

func (d *VMDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}

func (d *VMDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a virtual machine on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier for the VM.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the VM.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the VM.",
				Computed:    true,
			},
			"vcpus": schema.Int64Attribute{
				Description: "Number of virtual CPUs.",
				Computed:    true,
			},
			"cores": schema.Int64Attribute{
				Description: "Number of cores per socket.",
				Computed:    true,
			},
			"threads": schema.Int64Attribute{
				Description: "Number of threads per core.",
				Computed:    true,
			},
			"memory": schema.Int64Attribute{
				Description: "Memory in MiB.",
				Computed:    true,
			},
			"bootloader": schema.StringAttribute{
				Description: "Bootloader type.",
				Computed:    true,
			},
			"autostart": schema.BoolAttribute{
				Description: "Whether the VM autostarts.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Current status of the VM.",
				Computed:    true,
			},
			"uuid": schema.StringAttribute{
				Description: "VM UUID.",
				Computed:    true,
			},
		},
	}
}

func (d *VMDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData))
		return
	}
	d.client = client
}

func (d *VMDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config VMDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	var err error

	if !config.ID.IsNull() {
		err = d.client.GetInstance(ctx, "vm", config.ID.ValueInt64(), &result)
	} else if !config.Name.IsNull() {
		params := client.NewQueryParams().WithFilter("name", "=", config.Name.ValueString())
		var results []map[string]interface{}
		err = d.client.Query(ctx, "vm", params, &results)
		if err == nil && len(results) > 0 {
			result = results[0]
		} else if len(results) == 0 {
			resp.Diagnostics.AddError("VM Not Found", fmt.Sprintf("VM with name %s not found", config.Name.ValueString()))
			return
		}
	} else {
		resp.Diagnostics.AddError("Missing Identifier", "Either id or name must be specified")
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Error Reading VM", "Could not read VM: "+err.Error())
		return
	}

	config.ID = types.Int64Value(int64(result["id"].(float64)))
	config.Name = types.StringValue(result["name"].(string))

	if description, ok := result["description"].(string); ok {
		config.Description = types.StringValue(description)
	}
	if vcpus, ok := result["vcpus"].(float64); ok {
		config.VCPUs = types.Int64Value(int64(vcpus))
	}
	if cores, ok := result["cores"].(float64); ok {
		config.Cores = types.Int64Value(int64(cores))
	}
	if threads, ok := result["threads"].(float64); ok {
		config.Threads = types.Int64Value(int64(threads))
	}
	if memory, ok := result["memory"].(float64); ok {
		config.Memory = types.Int64Value(int64(memory))
	}
	if bootloader, ok := result["bootloader"].(string); ok {
		config.Bootloader = types.StringValue(bootloader)
	}
	if autostart, ok := result["autostart"].(bool); ok {
		config.Autostart = types.BoolValue(autostart)
	}
	if status, ok := result["status"].(map[string]interface{}); ok {
		if state, ok := status["state"].(string); ok {
			config.Status = types.StringValue(state)
		}
	}
	if uuid, ok := result["uuid"].(string); ok {
		config.UUID = types.StringValue(uuid)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}
