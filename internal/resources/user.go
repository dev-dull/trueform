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
	_ resource.Resource                = &UserResource{}
	_ resource.ResourceWithImportState = &UserResource{}
)

func NewUserResource() resource.Resource {
	return &UserResource{}
}

type UserResource struct {
	client *client.Client
}

type UserResourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	UID            types.Int64  `tfsdk:"uid"`
	Username       types.String `tfsdk:"username"`
	FullName       types.String `tfsdk:"full_name"`
	Email          types.String `tfsdk:"email"`
	Password       types.String `tfsdk:"password"`
	PasswordDisabled types.Bool `tfsdk:"password_disabled"`
	Group          types.Int64  `tfsdk:"group"`
	Groups         types.List   `tfsdk:"groups"`
	Home           types.String `tfsdk:"home"`
	HomeMode       types.String `tfsdk:"home_mode"`
	HomeCreate     types.Bool   `tfsdk:"home_create"`
	Shell          types.String `tfsdk:"shell"`
	SSHPubKey      types.String `tfsdk:"sshpubkey"`
	Locked         types.Bool   `tfsdk:"locked"`
	SMB            types.Bool   `tfsdk:"smb"`
	Sudo           types.Bool   `tfsdk:"sudo"`
	SudoNopasswd   types.Bool   `tfsdk:"sudo_nopasswd"`
	SudoCommands   types.List   `tfsdk:"sudo_commands"`
	Builtin        types.Bool   `tfsdk:"builtin"`
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a user on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier for the user.",
				Computed:    true,
			},
			"uid": schema.Int64Attribute{
				Description: "Unix user ID.",
				Optional:    true,
				Computed:    true,
			},
			"username": schema.StringAttribute{
				Description: "The username.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"full_name": schema.StringAttribute{
				Description: "Full name of the user.",
				Optional:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email address.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "User password.",
				Optional:    true,
				Sensitive:   true,
			},
			"password_disabled": schema.BoolAttribute{
				Description: "Disable password authentication.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"group": schema.Int64Attribute{
				Description: "Primary group ID.",
				Optional:    true,
				Computed:    true,
			},
			"groups": schema.ListAttribute{
				Description: "List of auxiliary group IDs.",
				Optional:    true,
				ElementType: types.Int64Type,
			},
			"home": schema.StringAttribute{
				Description: "Home directory path.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("/nonexistent"),
			},
			"home_mode": schema.StringAttribute{
				Description: "Home directory permissions.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("700"),
			},
			"home_create": schema.BoolAttribute{
				Description: "Create home directory if it doesn't exist.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"shell": schema.StringAttribute{
				Description: "Login shell.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("/usr/bin/zsh"),
			},
			"sshpubkey": schema.StringAttribute{
				Description: "SSH public key.",
				Optional:    true,
			},
			"locked": schema.BoolAttribute{
				Description: "Lock the user account.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"smb": schema.BoolAttribute{
				Description: "Enable SMB authentication for this user.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"sudo": schema.BoolAttribute{
				Description: "Allow sudo access.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"sudo_nopasswd": schema.BoolAttribute{
				Description: "Allow sudo without password.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"sudo_commands": schema.ListAttribute{
				Description: "List of allowed sudo commands.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"builtin": schema.BoolAttribute{
				Description: "Whether this is a built-in user.",
				Computed:    true,
			},
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating user", map[string]interface{}{
		"username": plan.Username.ValueString(),
	})

	createData := map[string]interface{}{
		"username":          plan.Username.ValueString(),
		"home":              plan.Home.ValueString(),
		"shell":             plan.Shell.ValueString(),
		"password_disabled": plan.PasswordDisabled.ValueBool(),
		"locked":            plan.Locked.ValueBool(),
		"smb":               plan.SMB.ValueBool(),
		"sudo":              plan.Sudo.ValueBool(),
		"sudo_nopasswd":     plan.SudoNopasswd.ValueBool(),
	}

	if !plan.UID.IsNull() {
		createData["uid"] = plan.UID.ValueInt64()
	}
	if !plan.FullName.IsNull() {
		createData["full_name"] = plan.FullName.ValueString()
	}
	if !plan.Email.IsNull() {
		createData["email"] = plan.Email.ValueString()
	}
	if !plan.Password.IsNull() && plan.Password.ValueString() != "" {
		createData["password"] = plan.Password.ValueString()
	}
	if !plan.Group.IsNull() {
		createData["group"] = plan.Group.ValueInt64()
	}
	if !plan.Groups.IsNull() {
		var groups []int64
		diags = plan.Groups.ElementsAs(ctx, &groups, false)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			createData["groups"] = groups
		}
	}
	if !plan.HomeMode.IsNull() {
		createData["home_mode"] = plan.HomeMode.ValueString()
	}
	if !plan.HomeCreate.IsNull() {
		createData["home_create"] = plan.HomeCreate.ValueBool()
	}
	if !plan.SSHPubKey.IsNull() {
		createData["sshpubkey"] = plan.SSHPubKey.ValueString()
	}
	if !plan.SudoCommands.IsNull() {
		var commands []string
		diags = plan.SudoCommands.ElementsAs(ctx, &commands, false)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			createData["sudo_commands"] = commands
		}
	}

	var result map[string]interface{}
	err := r.client.Create(ctx, "user", createData, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating User", "Could not create user: "+err.Error())
		return
	}

	userID := int64(result["id"].(float64))
	if err := r.readUser(ctx, userID, &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading User", "Could not read user after creation: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readUser(ctx, state.ID.ValueInt64(), &state); err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading User", "Could not read user: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state UserResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateData := map[string]interface{}{}

	if !plan.FullName.Equal(state.FullName) {
		if plan.FullName.IsNull() {
			updateData["full_name"] = ""
		} else {
			updateData["full_name"] = plan.FullName.ValueString()
		}
	}
	if !plan.Email.Equal(state.Email) {
		if plan.Email.IsNull() {
			updateData["email"] = ""
		} else {
			updateData["email"] = plan.Email.ValueString()
		}
	}
	if !plan.Password.Equal(state.Password) && !plan.Password.IsNull() && plan.Password.ValueString() != "" {
		updateData["password"] = plan.Password.ValueString()
	}
	if !plan.PasswordDisabled.Equal(state.PasswordDisabled) {
		updateData["password_disabled"] = plan.PasswordDisabled.ValueBool()
	}
	if !plan.Group.Equal(state.Group) {
		updateData["group"] = plan.Group.ValueInt64()
	}
	if !plan.Groups.Equal(state.Groups) {
		var groups []int64
		if !plan.Groups.IsNull() {
			diags = plan.Groups.ElementsAs(ctx, &groups, false)
			resp.Diagnostics.Append(diags...)
		}
		updateData["groups"] = groups
	}
	if !plan.Home.Equal(state.Home) {
		updateData["home"] = plan.Home.ValueString()
	}
	if !plan.HomeMode.Equal(state.HomeMode) {
		updateData["home_mode"] = plan.HomeMode.ValueString()
	}
	if !plan.Shell.Equal(state.Shell) {
		updateData["shell"] = plan.Shell.ValueString()
	}
	if !plan.SSHPubKey.Equal(state.SSHPubKey) {
		if plan.SSHPubKey.IsNull() {
			updateData["sshpubkey"] = ""
		} else {
			updateData["sshpubkey"] = plan.SSHPubKey.ValueString()
		}
	}
	if !plan.Locked.Equal(state.Locked) {
		updateData["locked"] = plan.Locked.ValueBool()
	}
	if !plan.SMB.Equal(state.SMB) {
		updateData["smb"] = plan.SMB.ValueBool()
	}
	if !plan.Sudo.Equal(state.Sudo) {
		updateData["sudo"] = plan.Sudo.ValueBool()
	}
	if !plan.SudoNopasswd.Equal(state.SudoNopasswd) {
		updateData["sudo_nopasswd"] = plan.SudoNopasswd.ValueBool()
	}
	if !plan.SudoCommands.Equal(state.SudoCommands) {
		var commands []string
		if !plan.SudoCommands.IsNull() {
			diags = plan.SudoCommands.ElementsAs(ctx, &commands, false)
			resp.Diagnostics.Append(diags...)
		}
		updateData["sudo_commands"] = commands
	}

	if len(updateData) > 0 {
		var result map[string]interface{}
		err := r.client.Update(ctx, "user", state.ID.ValueInt64(), updateData, &result)
		if err != nil {
			resp.Diagnostics.AddError("Error Updating User", "Could not update user: "+err.Error())
			return
		}
	}

	if err := r.readUser(ctx, state.ID.ValueInt64(), &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading User", "Could not read user after update: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "user", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting User", "Could not delete user: "+err.Error())
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *UserResource) readUser(ctx context.Context, id int64, model *UserResourceModel) error {
	var result map[string]interface{}
	err := r.client.GetInstance(ctx, "user", id, &result)
	if err != nil {
		return err
	}

	model.ID = types.Int64Value(int64(result["id"].(float64)))
	model.UID = types.Int64Value(int64(result["uid"].(float64)))
	model.Username = types.StringValue(result["username"].(string))

	if fullName, ok := result["full_name"].(string); ok {
		model.FullName = types.StringValue(fullName)
	}
	if email, ok := result["email"].(string); ok {
		model.Email = types.StringValue(email)
	}
	if passwordDisabled, ok := result["password_disabled"].(bool); ok {
		model.PasswordDisabled = types.BoolValue(passwordDisabled)
	}
	if group, ok := result["group"].(map[string]interface{}); ok {
		if gid, ok := group["id"].(float64); ok {
			model.Group = types.Int64Value(int64(gid))
		}
	}
	if groups, ok := result["groups"].([]interface{}); ok {
		groupIDs := make([]int64, len(groups))
		for i, g := range groups {
			if gmap, ok := g.(map[string]interface{}); ok {
				if gid, ok := gmap["id"].(float64); ok {
					groupIDs[i] = int64(gid)
				}
			}
		}
		groupValues, diags := types.ListValueFrom(ctx, types.Int64Type, groupIDs)
		if !diags.HasError() {
			model.Groups = groupValues
		}
	}
	if home, ok := result["home"].(string); ok {
		model.Home = types.StringValue(home)
	}
	if shell, ok := result["shell"].(string); ok {
		model.Shell = types.StringValue(shell)
	}
	if sshpubkey, ok := result["sshpubkey"].(string); ok {
		model.SSHPubKey = types.StringValue(sshpubkey)
	}
	if locked, ok := result["locked"].(bool); ok {
		model.Locked = types.BoolValue(locked)
	}
	if smb, ok := result["smb"].(bool); ok {
		model.SMB = types.BoolValue(smb)
	}
	if sudo, ok := result["sudo"].(bool); ok {
		model.Sudo = types.BoolValue(sudo)
	}
	if sudoNopasswd, ok := result["sudo_nopasswd"].(bool); ok {
		model.SudoNopasswd = types.BoolValue(sudoNopasswd)
	}
	if sudoCommands, ok := result["sudo_commands"].([]interface{}); ok {
		commands := make([]string, len(sudoCommands))
		for i, cmd := range sudoCommands {
			commands[i] = cmd.(string)
		}
		cmdValues, diags := types.ListValueFrom(ctx, types.StringType, commands)
		if !diags.HasError() {
			model.SudoCommands = cmdValues
		}
	}
	if builtin, ok := result["builtin"].(bool); ok {
		model.Builtin = types.BoolValue(builtin)
	}

	return nil
}
