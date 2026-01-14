package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/trueform/terraform-provider-trueform/internal/client"
)

var _ datasource.DataSource = &DatasetDataSource{}

func NewDatasetDataSource() datasource.DataSource {
	return &DatasetDataSource{}
}

type DatasetDataSource struct {
	client *client.Client
}

type DatasetDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Pool            types.String `tfsdk:"pool"`
	Type            types.String `tfsdk:"type"`
	Compression     types.String `tfsdk:"compression"`
	Atime           types.String `tfsdk:"atime"`
	Deduplication   types.String `tfsdk:"deduplication"`
	Quota           types.Int64  `tfsdk:"quota"`
	Used            types.Int64  `tfsdk:"used"`
	Available       types.Int64  `tfsdk:"available"`
	Mountpoint      types.String `tfsdk:"mountpoint"`
	Encrypted       types.Bool   `tfsdk:"encrypted"`
	KeyLoaded       types.Bool   `tfsdk:"key_loaded"`
}

func (d *DatasetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataset"
}

func (d *DatasetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a ZFS dataset on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The full path identifier for the dataset.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the dataset.",
				Computed:    true,
			},
			"pool": schema.StringAttribute{
				Description: "The pool containing the dataset.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Dataset type (FILESYSTEM or VOLUME).",
				Computed:    true,
			},
			"compression": schema.StringAttribute{
				Description: "Compression algorithm.",
				Computed:    true,
			},
			"atime": schema.StringAttribute{
				Description: "Access time setting.",
				Computed:    true,
			},
			"deduplication": schema.StringAttribute{
				Description: "Deduplication setting.",
				Computed:    true,
			},
			"quota": schema.Int64Attribute{
				Description: "Quota in bytes.",
				Computed:    true,
			},
			"used": schema.Int64Attribute{
				Description: "Space used in bytes.",
				Computed:    true,
			},
			"available": schema.Int64Attribute{
				Description: "Space available in bytes.",
				Computed:    true,
			},
			"mountpoint": schema.StringAttribute{
				Description: "Mount point path.",
				Computed:    true,
			},
			"encrypted": schema.BoolAttribute{
				Description: "Whether the dataset is encrypted.",
				Computed:    true,
			},
			"key_loaded": schema.BoolAttribute{
				Description: "Whether the encryption key is loaded.",
				Computed:    true,
			},
		},
	}
}

func (d *DatasetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DatasetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DatasetDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	err := d.client.GetInstance(ctx, "pool.dataset", config.ID.ValueString(), &result)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dataset", "Could not read dataset: "+err.Error())
		return
	}

	config.ID = types.StringValue(result["id"].(string))
	config.Name = types.StringValue(result["name"].(string))
	config.Type = types.StringValue(result["type"].(string))

	// Extract pool from name
	name := result["name"].(string)
	for i, c := range name {
		if c == '/' {
			config.Pool = types.StringValue(name[:i])
			break
		}
	}
	if config.Pool.IsNull() {
		config.Pool = types.StringValue(name)
	}

	if compression, ok := result["compression"].(map[string]interface{}); ok {
		if value, ok := compression["value"].(string); ok {
			config.Compression = types.StringValue(value)
		}
	}
	if atime, ok := result["atime"].(map[string]interface{}); ok {
		if value, ok := atime["value"].(string); ok {
			config.Atime = types.StringValue(value)
		}
	}
	if dedup, ok := result["deduplication"].(map[string]interface{}); ok {
		if value, ok := dedup["value"].(string); ok {
			config.Deduplication = types.StringValue(value)
		}
	}
	if quota, ok := result["quota"].(map[string]interface{}); ok {
		if parsed, ok := quota["parsed"].(float64); ok {
			config.Quota = types.Int64Value(int64(parsed))
		}
	}
	if used, ok := result["used"].(map[string]interface{}); ok {
		if parsed, ok := used["parsed"].(float64); ok {
			config.Used = types.Int64Value(int64(parsed))
		}
	}
	if available, ok := result["available"].(map[string]interface{}); ok {
		if parsed, ok := available["parsed"].(float64); ok {
			config.Available = types.Int64Value(int64(parsed))
		}
	}
	if mountpoint, ok := result["mountpoint"].(string); ok {
		config.Mountpoint = types.StringValue(mountpoint)
	}
	if encrypted, ok := result["encrypted"].(bool); ok {
		config.Encrypted = types.BoolValue(encrypted)
	}
	if keyLoaded, ok := result["key_loaded"].(bool); ok {
		config.KeyLoaded = types.BoolValue(keyLoaded)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}
