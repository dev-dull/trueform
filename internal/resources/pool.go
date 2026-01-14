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
	_ resource.Resource                = &PoolResource{}
	_ resource.ResourceWithImportState = &PoolResource{}
)

func NewPoolResource() resource.Resource {
	return &PoolResource{}
}

type PoolResource struct {
	client *client.Client
}

type PoolResourceModel struct {
	ID                types.Int64  `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Topology          types.List   `tfsdk:"topology"`
	Encryption        types.Bool   `tfsdk:"encryption"`
	EncryptionOptions types.Object `tfsdk:"encryption_options"`
	Deduplication     types.String `tfsdk:"deduplication"`
	Checksum          types.String `tfsdk:"checksum"`
	Status            types.String `tfsdk:"status"`
	Healthy           types.Bool   `tfsdk:"healthy"`
	Path              types.String `tfsdk:"path"`
	Size              types.Int64  `tfsdk:"size"`
	Free              types.Int64  `tfsdk:"free"`
	Allocated         types.Int64  `tfsdk:"allocated"`
}

type TopologyVDev struct {
	Type  types.String `tfsdk:"type"`
	Disks types.List   `tfsdk:"disks"`
}

func (r *PoolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pool"
}

func (r *PoolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a ZFS pool on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier for the pool.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the pool.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"encryption": schema.BoolAttribute{
				Description: "Whether to enable encryption on the pool.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"encryption_options": schema.SingleNestedAttribute{
				Description: "Encryption options for the pool.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"algorithm": schema.StringAttribute{
						Description: "Encryption algorithm (e.g., AES-256-GCM).",
						Optional:    true,
					},
					"passphrase": schema.StringAttribute{
						Description: "Encryption passphrase.",
						Optional:    true,
						Sensitive:   true,
					},
				},
			},
			"deduplication": schema.StringAttribute{
				Description: "Deduplication setting (on, off, verify).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("off"),
			},
			"checksum": schema.StringAttribute{
				Description: "Checksum algorithm.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("on"),
			},
			"status": schema.StringAttribute{
				Description: "The status of the pool (ONLINE, DEGRADED, FAULTED, etc.).",
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
			"topology": schema.ListNestedAttribute{
				Description: "The topology configuration for the pool.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "The vdev type (data, log, cache, spare, special, dedup).",
							Required:    true,
						},
						"disks": schema.ListAttribute{
							Description: "List of disk identifiers for this vdev.",
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (r *PoolResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PoolResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating pool", map[string]interface{}{
		"name": plan.Name.ValueString(),
	})

	// Build topology structure for API
	var topologyVDevs []TopologyVDev
	diags = plan.Topology.ElementsAs(ctx, &topologyVDevs, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	topology := make(map[string]interface{})
	for _, vdev := range topologyVDevs {
		var disks []string
		diags = vdev.Disks.ElementsAs(ctx, &disks, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		vdevType := vdev.Type.ValueString()
		if topology[vdevType] == nil {
			topology[vdevType] = []map[string]interface{}{}
		}

		vdevData := map[string]interface{}{
			"type":  "STRIPE", // Default, will be determined by number of disks
			"disks": disks,
		}

		// Determine vdev type based on disk count
		switch len(disks) {
		case 1:
			vdevData["type"] = "STRIPE"
		case 2:
			vdevData["type"] = "MIRROR"
		default:
			vdevData["type"] = "RAIDZ1"
		}

		topology[vdevType] = append(topology[vdevType].([]map[string]interface{}), vdevData)
	}

	createData := map[string]interface{}{
		"name":     plan.Name.ValueString(),
		"topology": topology,
	}

	if !plan.Encryption.IsNull() && plan.Encryption.ValueBool() {
		createData["encryption"] = true
		// Add encryption options if specified
	}

	if !plan.Deduplication.IsNull() {
		createData["deduplication"] = plan.Deduplication.ValueString()
	}

	var result map[string]interface{}
	err := r.client.Create(ctx, "pool", createData, &result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Pool",
			"Could not create pool: "+err.Error(),
		)
		return
	}

	// Get the created pool to populate computed fields
	poolID := int64(result["id"].(float64))
	if err := r.readPool(ctx, poolID, &plan); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Pool",
			"Could not read pool after creation: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PoolResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readPool(ctx, state.ID.ValueInt64(), &state); err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Pool",
			"Could not read pool: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *PoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PoolResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PoolResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating pool", map[string]interface{}{
		"id": state.ID.ValueInt64(),
	})

	// Pools have limited update capabilities in TrueNAS
	// Most changes require recreation
	updateData := map[string]interface{}{}

	if !plan.Checksum.Equal(state.Checksum) {
		updateData["checksum"] = plan.Checksum.ValueString()
	}

	if len(updateData) > 0 {
		var result map[string]interface{}
		err := r.client.Update(ctx, "pool", state.ID.ValueInt64(), updateData, &result)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Pool",
				"Could not update pool: "+err.Error(),
			)
			return
		}
	}

	// Read the updated pool
	if err := r.readPool(ctx, state.ID.ValueInt64(), &plan); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Pool",
			"Could not read pool after update: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PoolResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting pool", map[string]interface{}{
		"id": state.ID.ValueInt64(),
	})

	// Export and destroy the pool
	err := r.client.Call(ctx, "pool.export", []interface{}{
		state.ID.ValueInt64(),
		map[string]interface{}{
			"destroy": true,
		},
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Pool",
			"Could not delete pool: "+err.Error(),
		)
		return
	}
}

func (r *PoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *PoolResource) readPool(ctx context.Context, id int64, model *PoolResourceModel) error {
	var result map[string]interface{}
	err := r.client.GetInstance(ctx, "pool", id, &result)
	if err != nil {
		return err
	}

	model.ID = types.Int64Value(int64(result["id"].(float64)))
	model.Name = types.StringValue(result["name"].(string))
	model.Status = types.StringValue(result["status"].(string))
	model.Healthy = types.BoolValue(result["healthy"].(bool))
	model.Path = types.StringValue(result["path"].(string))

	if size, ok := result["size"].(float64); ok {
		model.Size = types.Int64Value(int64(size))
	}
	if free, ok := result["free"].(float64); ok {
		model.Free = types.Int64Value(int64(free))
	}
	if allocated, ok := result["allocated"].(float64); ok {
		model.Allocated = types.Int64Value(int64(allocated))
	}

	return nil
}
