package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/trueform/terraform-provider-trueform/internal/client"
)

var _ datasource.DataSource = &PoolDataSource{}

func NewPoolDataSource() datasource.DataSource {
	return &PoolDataSource{}
}

type PoolDataSource struct {
	client *client.Client
}

type PoolDataSourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Status    types.String `tfsdk:"status"`
	Healthy   types.Bool   `tfsdk:"healthy"`
	Path      types.String `tfsdk:"path"`
	Size      types.Int64  `tfsdk:"size"`
	Free      types.Int64  `tfsdk:"free"`
	Allocated types.Int64  `tfsdk:"allocated"`
	Fragmentation types.Int64 `tfsdk:"fragmentation"`
}

func (d *PoolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pool"
}

func (d *PoolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a ZFS pool on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier for the pool.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the pool.",
				Optional:    true,
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the pool.",
				Computed:    true,
			},
			"healthy": schema.BoolAttribute{
				Description: "Whether the pool is healthy.",
				Computed:    true,
			},
			"path": schema.StringAttribute{
				Description: "The mount path of the pool.",
				Computed:    true,
			},
			"size": schema.Int64Attribute{
				Description: "Total size of the pool in bytes.",
				Computed:    true,
			},
			"free": schema.Int64Attribute{
				Description: "Free space in the pool in bytes.",
				Computed:    true,
			},
			"allocated": schema.Int64Attribute{
				Description: "Allocated space in the pool in bytes.",
				Computed:    true,
			},
			"fragmentation": schema.Int64Attribute{
				Description: "Pool fragmentation percentage.",
				Computed:    true,
			},
		},
	}
}

func (d *PoolDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config PoolDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	var err error

	if !config.ID.IsNull() {
		err = d.client.GetInstance(ctx, "pool", config.ID.ValueInt64(), &result)
	} else if !config.Name.IsNull() {
		// Query by name
		params := client.NewQueryParams().WithFilter("name", "=", config.Name.ValueString())
		var results []map[string]interface{}
		err = d.client.Query(ctx, "pool", params, &results)
		if err == nil && len(results) > 0 {
			result = results[0]
		} else if len(results) == 0 {
			resp.Diagnostics.AddError("Pool Not Found", fmt.Sprintf("Pool with name %s not found", config.Name.ValueString()))
			return
		}
	} else {
		resp.Diagnostics.AddError("Missing Identifier", "Either id or name must be specified")
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Error Reading Pool", "Could not read pool: "+err.Error())
		return
	}

	config.ID = types.Int64Value(int64(result["id"].(float64)))
	config.Name = types.StringValue(result["name"].(string))
	config.Status = types.StringValue(result["status"].(string))
	config.Healthy = types.BoolValue(result["healthy"].(bool))
	config.Path = types.StringValue(result["path"].(string))

	if size, ok := result["size"].(float64); ok {
		config.Size = types.Int64Value(int64(size))
	}
	if free, ok := result["free"].(float64); ok {
		config.Free = types.Int64Value(int64(free))
	}
	if allocated, ok := result["allocated"].(float64); ok {
		config.Allocated = types.Int64Value(int64(allocated))
	}
	if fragmentation, ok := result["fragmentation"].(float64); ok {
		config.Fragmentation = types.Int64Value(int64(fragmentation))
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}
