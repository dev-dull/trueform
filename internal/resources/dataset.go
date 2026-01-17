package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/trueform/terraform-provider-trueform/internal/client"
)

var (
	_ resource.Resource                = &DatasetResource{}
	_ resource.ResourceWithImportState = &DatasetResource{}
)

func NewDatasetResource() resource.Resource {
	return &DatasetResource{}
}

type DatasetResource struct {
	client *client.Client
}

type DatasetResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Pool            types.String `tfsdk:"pool"`
	Type            types.String `tfsdk:"type"`
	Comments        types.String `tfsdk:"comments"`
	Compression     types.String `tfsdk:"compression"`
	Atime           types.String `tfsdk:"atime"`
	Deduplication   types.String `tfsdk:"deduplication"`
	Quota           types.Int64  `tfsdk:"quota"`
	QuotaWarning    types.Int64  `tfsdk:"quota_warning"`
	QuotaCritical   types.Int64  `tfsdk:"quota_critical"`
	Refquota        types.Int64  `tfsdk:"refquota"`
	Reservation     types.Int64  `tfsdk:"reservation"`
	Refreservation  types.Int64  `tfsdk:"refreservation"`
	Copies          types.Int64  `tfsdk:"copies"`
	Snapdir         types.String `tfsdk:"snapdir"`
	Readonly        types.String `tfsdk:"readonly"`
	Recordsize      types.String `tfsdk:"recordsize"`
	Casesensitivity types.String `tfsdk:"casesensitivity"`
	Aclmode         types.String `tfsdk:"aclmode"`
	Acltype         types.String `tfsdk:"acltype"`
	ShareType       types.String `tfsdk:"share_type"`
	ManagedBy       types.String `tfsdk:"managed_by"`
	Mountpoint      types.String `tfsdk:"mountpoint"`
	Encrypted       types.Bool   `tfsdk:"encrypted"`
	EncryptionRoot  types.String `tfsdk:"encryption_root"`
	KeyLoaded       types.Bool   `tfsdk:"key_loaded"`
	Used            types.Int64  `tfsdk:"used"`
	Available       types.Int64  `tfsdk:"available"`
}

func (r *DatasetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataset"
}

func (r *DatasetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a ZFS dataset on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the dataset (full path).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the dataset (relative to pool).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pool": schema.StringAttribute{
				Description: "The pool where the dataset resides.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Dataset type (FILESYSTEM or VOLUME).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("FILESYSTEM"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"comments": schema.StringAttribute{
				Description: "Comments for the dataset.",
				Optional:    true,
			},
			"compression": schema.StringAttribute{
				Description: "Compression algorithm (OFF, LZ4, GZIP, ZSTD, etc.).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("LZ4"),
			},
			"atime": schema.StringAttribute{
				Description: "Access time update setting (ON, OFF).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("OFF"),
			},
			"deduplication": schema.StringAttribute{
				Description: "Deduplication setting (ON, OFF, VERIFY).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("OFF"),
			},
			"quota": schema.Int64Attribute{
				Description: "Quota in bytes (0 for unlimited).",
				Optional:    true,
			},
			"quota_warning": schema.Int64Attribute{
				Description: "Quota warning threshold percentage.",
				Optional:    true,
			},
			"quota_critical": schema.Int64Attribute{
				Description: "Quota critical threshold percentage.",
				Optional:    true,
			},
			"refquota": schema.Int64Attribute{
				Description: "Reference quota in bytes.",
				Optional:    true,
			},
			"reservation": schema.Int64Attribute{
				Description: "Reservation in bytes.",
				Optional:    true,
			},
			"refreservation": schema.Int64Attribute{
				Description: "Reference reservation in bytes.",
				Optional:    true,
			},
			"copies": schema.Int64Attribute{
				Description: "Number of copies to store.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
			},
			"snapdir": schema.StringAttribute{
				Description: "Snapshot directory visibility (VISIBLE, HIDDEN, DISABLED).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("HIDDEN"),
			},
			"readonly": schema.StringAttribute{
				Description: "Read-only setting (ON, OFF).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("OFF"),
			},
			"recordsize": schema.StringAttribute{
				Description: "Record size (e.g., 128K, 1M).",
				Optional:    true,
				Computed:    true,
			},
			"casesensitivity": schema.StringAttribute{
				Description: "Case sensitivity (sensitive, insensitive, mixed).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"aclmode": schema.StringAttribute{
				Description: "ACL mode (passthrough, restricted, discard).",
				Optional:    true,
				Computed:    true,
			},
			"acltype": schema.StringAttribute{
				Description: "ACL type (off, nfsv4, posix).",
				Optional:    true,
				Computed:    true,
			},
			"share_type": schema.StringAttribute{
				Description: "Share type (GENERIC, SMB, NFS, MULTIPROTOCOL, APPS).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("GENERIC"),
			},
			"managed_by": schema.StringAttribute{
				Description: "What manages this dataset.",
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
			"encryption_root": schema.StringAttribute{
				Description: "Encryption root dataset.",
				Computed:    true,
			},
			"key_loaded": schema.BoolAttribute{
				Description: "Whether the encryption key is loaded.",
				Computed:    true,
			},
			"used": schema.Int64Attribute{
				Description: "Space used by the dataset in bytes.",
				Computed:    true,
			},
			"available": schema.Int64Attribute{
				Description: "Space available to the dataset in bytes.",
				Computed:    true,
			},
		},
	}
}

func (r *DatasetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DatasetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DatasetResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build full dataset path
	datasetPath := plan.Pool.ValueString() + "/" + plan.Name.ValueString()

	tflog.Debug(ctx, "Creating dataset", map[string]interface{}{
		"path": datasetPath,
	})

	createData := map[string]interface{}{
		"name": datasetPath,
		"type": plan.Type.ValueString(),
	}

	if !plan.Comments.IsNull() {
		createData["comments"] = plan.Comments.ValueString()
	}
	if !plan.Compression.IsNull() {
		createData["compression"] = plan.Compression.ValueString()
	}
	if !plan.Atime.IsNull() {
		createData["atime"] = plan.Atime.ValueString()
	}
	if !plan.Deduplication.IsNull() {
		createData["deduplication"] = plan.Deduplication.ValueString()
	}
	if !plan.Quota.IsNull() {
		createData["quota"] = plan.Quota.ValueInt64()
	}
	if !plan.QuotaWarning.IsNull() {
		createData["quota_warning"] = plan.QuotaWarning.ValueInt64()
	}
	if !plan.QuotaCritical.IsNull() {
		createData["quota_critical"] = plan.QuotaCritical.ValueInt64()
	}
	if !plan.Refquota.IsNull() {
		createData["refquota"] = plan.Refquota.ValueInt64()
	}
	if !plan.Reservation.IsNull() {
		createData["reservation"] = plan.Reservation.ValueInt64()
	}
	if !plan.Refreservation.IsNull() {
		createData["refreservation"] = plan.Refreservation.ValueInt64()
	}
	if !plan.Copies.IsNull() {
		createData["copies"] = plan.Copies.ValueInt64()
	}
	if !plan.Snapdir.IsNull() && !plan.Snapdir.IsUnknown() {
		createData["snapdir"] = plan.Snapdir.ValueString()
	}
	if !plan.Readonly.IsNull() && !plan.Readonly.IsUnknown() {
		createData["readonly"] = plan.Readonly.ValueString()
	}
	if !plan.Recordsize.IsNull() && !plan.Recordsize.IsUnknown() {
		createData["recordsize"] = plan.Recordsize.ValueString()
	}
	if !plan.Casesensitivity.IsNull() && !plan.Casesensitivity.IsUnknown() {
		createData["casesensitivity"] = plan.Casesensitivity.ValueString()
	}
	if !plan.Aclmode.IsNull() && !plan.Aclmode.IsUnknown() {
		createData["aclmode"] = plan.Aclmode.ValueString()
	}
	if !plan.Acltype.IsNull() && !plan.Acltype.IsUnknown() {
		createData["acltype"] = plan.Acltype.ValueString()
	}
	if !plan.ShareType.IsNull() && !plan.ShareType.IsUnknown() {
		createData["share_type"] = plan.ShareType.ValueString()
	}

	var result map[string]interface{}
	err := r.client.Create(ctx, "pool.dataset", createData, &result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Dataset",
			"Could not create dataset: "+err.Error(),
		)
		return
	}

	// Read the created dataset
	if err := r.readDataset(ctx, datasetPath, &plan); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Dataset",
			"Could not read dataset after creation: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DatasetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DatasetResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readDataset(ctx, state.ID.ValueString(), &state); err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Dataset",
			"Could not read dataset: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *DatasetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DatasetResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state DatasetResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating dataset", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	updateData := map[string]interface{}{}

	if !plan.Comments.Equal(state.Comments) {
		if plan.Comments.IsNull() {
			updateData["comments"] = ""
		} else {
			updateData["comments"] = plan.Comments.ValueString()
		}
	}
	if !plan.Compression.Equal(state.Compression) {
		updateData["compression"] = plan.Compression.ValueString()
	}
	if !plan.Atime.Equal(state.Atime) {
		updateData["atime"] = plan.Atime.ValueString()
	}
	if !plan.Deduplication.Equal(state.Deduplication) {
		updateData["deduplication"] = plan.Deduplication.ValueString()
	}
	if !plan.Quota.Equal(state.Quota) {
		updateData["quota"] = plan.Quota.ValueInt64()
	}
	if !plan.QuotaWarning.Equal(state.QuotaWarning) {
		updateData["quota_warning"] = plan.QuotaWarning.ValueInt64()
	}
	if !plan.QuotaCritical.Equal(state.QuotaCritical) {
		updateData["quota_critical"] = plan.QuotaCritical.ValueInt64()
	}
	if !plan.Refquota.Equal(state.Refquota) {
		updateData["refquota"] = plan.Refquota.ValueInt64()
	}
	if !plan.Reservation.Equal(state.Reservation) {
		updateData["reservation"] = plan.Reservation.ValueInt64()
	}
	if !plan.Refreservation.Equal(state.Refreservation) {
		updateData["refreservation"] = plan.Refreservation.ValueInt64()
	}
	if !plan.Copies.Equal(state.Copies) {
		updateData["copies"] = plan.Copies.ValueInt64()
	}
	if !plan.Snapdir.Equal(state.Snapdir) {
		updateData["snapdir"] = plan.Snapdir.ValueString()
	}
	if !plan.Readonly.Equal(state.Readonly) {
		updateData["readonly"] = plan.Readonly.ValueString()
	}
	// Only update computed fields if explicitly configured with valid values
	if !plan.Recordsize.Equal(state.Recordsize) && !plan.Recordsize.IsNull() && !plan.Recordsize.IsUnknown() && plan.Recordsize.ValueString() != "" {
		updateData["recordsize"] = plan.Recordsize.ValueString()
	}
	if !plan.Aclmode.Equal(state.Aclmode) && !plan.Aclmode.IsNull() && !plan.Aclmode.IsUnknown() && plan.Aclmode.ValueString() != "" {
		updateData["aclmode"] = plan.Aclmode.ValueString()
	}
	if !plan.Acltype.Equal(state.Acltype) && !plan.Acltype.IsNull() && !plan.Acltype.IsUnknown() && plan.Acltype.ValueString() != "" {
		updateData["acltype"] = plan.Acltype.ValueString()
	}
	if !plan.ShareType.Equal(state.ShareType) && !plan.ShareType.IsNull() && !plan.ShareType.IsUnknown() && plan.ShareType.ValueString() != "" {
		updateData["share_type"] = plan.ShareType.ValueString()
	}

	if len(updateData) > 0 {
		var result map[string]interface{}
		err := r.client.Update(ctx, "pool.dataset", state.ID.ValueString(), updateData, &result)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Dataset",
				"Could not update dataset: "+err.Error(),
			)
			return
		}
	}

	// Read the updated dataset
	if err := r.readDataset(ctx, state.ID.ValueString(), &plan); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Dataset",
			"Could not read dataset after update: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DatasetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DatasetResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting dataset", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	err := r.client.Delete(ctx, "pool.dataset", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Dataset",
			"Could not delete dataset: "+err.Error(),
		)
		return
	}
}

func (r *DatasetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *DatasetResource) readDataset(ctx context.Context, id string, model *DatasetResourceModel) error {
	var result map[string]interface{}
	err := r.client.GetInstance(ctx, "pool.dataset", id, &result)
	if err != nil {
		return err
	}

	model.ID = types.StringValue(result["id"].(string))

	// Extract pool and name from the full path
	fullName := result["name"].(string)
	parts := strings.SplitN(fullName, "/", 2)
	model.Pool = types.StringValue(parts[0])
	if len(parts) > 1 {
		model.Name = types.StringValue(parts[1])
	} else {
		model.Name = types.StringValue("")
	}

	model.Type = types.StringValue(result["type"].(string))

	if comments, ok := result["comments"].(string); ok {
		model.Comments = types.StringValue(comments)
	}
	if compression, ok := result["compression"].(map[string]interface{}); ok {
		if value, ok := compression["value"].(string); ok {
			model.Compression = types.StringValue(value)
		}
	}
	if atime, ok := result["atime"].(map[string]interface{}); ok {
		if value, ok := atime["value"].(string); ok {
			model.Atime = types.StringValue(value)
		}
	}
	if dedup, ok := result["deduplication"].(map[string]interface{}); ok {
		if value, ok := dedup["value"].(string); ok {
			model.Deduplication = types.StringValue(value)
		}
	}
	if quota, ok := result["quota"].(map[string]interface{}); ok {
		if parsed, ok := quota["parsed"].(float64); ok {
			model.Quota = types.Int64Value(int64(parsed))
		}
	}
	if refquota, ok := result["refquota"].(map[string]interface{}); ok {
		if parsed, ok := refquota["parsed"].(float64); ok {
			model.Refquota = types.Int64Value(int64(parsed))
		}
	}
	if reservation, ok := result["reservation"].(map[string]interface{}); ok {
		if parsed, ok := reservation["parsed"].(float64); ok {
			model.Reservation = types.Int64Value(int64(parsed))
		}
	}
	if refreservation, ok := result["refreservation"].(map[string]interface{}); ok {
		if parsed, ok := refreservation["parsed"].(float64); ok {
			model.Refreservation = types.Int64Value(int64(parsed))
		}
	}
	if copies, ok := result["copies"].(map[string]interface{}); ok {
		if value, ok := copies["value"].(string); ok {
			// Parse string to int
			var c int64
			fmt.Sscanf(value, "%d", &c)
			model.Copies = types.Int64Value(c)
		}
	}
	if snapdir, ok := result["snapdir"].(map[string]interface{}); ok {
		if value, ok := snapdir["value"].(string); ok {
			model.Snapdir = types.StringValue(value)
		}
	}
	if readonly, ok := result["readonly"].(map[string]interface{}); ok {
		if value, ok := readonly["value"].(string); ok {
			model.Readonly = types.StringValue(value)
		}
	}
	if recordsize, ok := result["recordsize"].(map[string]interface{}); ok {
		if value, ok := recordsize["value"].(string); ok {
			model.Recordsize = types.StringValue(value)
		}
	}
	if casesens, ok := result["casesensitivity"].(map[string]interface{}); ok {
		if value, ok := casesens["value"].(string); ok {
			model.Casesensitivity = types.StringValue(value)
		}
	}
	if aclmode, ok := result["aclmode"].(map[string]interface{}); ok {
		if value, ok := aclmode["value"].(string); ok {
			model.Aclmode = types.StringValue(value)
		}
	}
	if acltype, ok := result["acltype"].(map[string]interface{}); ok {
		if value, ok := acltype["value"].(string); ok {
			model.Acltype = types.StringValue(value)
		}
	}
	if managedBy, ok := result["managedby"].(map[string]interface{}); ok {
		if value, ok := managedBy["value"].(string); ok && value != "" {
			model.ManagedBy = types.StringValue(value)
		} else {
			model.ManagedBy = types.StringNull()
		}
	} else {
		model.ManagedBy = types.StringNull()
	}
	if mountpoint, ok := result["mountpoint"].(string); ok {
		model.Mountpoint = types.StringValue(mountpoint)
	}
	if encrypted, ok := result["encrypted"].(bool); ok {
		model.Encrypted = types.BoolValue(encrypted)
	}
	if encryptionRoot, ok := result["encryption_root"].(string); ok && encryptionRoot != "" {
		model.EncryptionRoot = types.StringValue(encryptionRoot)
	} else {
		model.EncryptionRoot = types.StringNull()
	}
	if keyLoaded, ok := result["key_loaded"].(bool); ok {
		model.KeyLoaded = types.BoolValue(keyLoaded)
	}
	if used, ok := result["used"].(map[string]interface{}); ok {
		if parsed, ok := used["parsed"].(float64); ok {
			model.Used = types.Int64Value(int64(parsed))
		}
	}
	if available, ok := result["available"].(map[string]interface{}); ok {
		if parsed, ok := available["parsed"].(float64); ok {
			model.Available = types.Int64Value(int64(parsed))
		}
	}

	return nil
}
