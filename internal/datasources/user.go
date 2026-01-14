package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/trueform/terraform-provider-trueform/internal/client"
)

var _ datasource.DataSource = &UserDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

type UserDataSource struct {
	client *client.Client
}

type UserDataSourceModel struct {
	ID       types.Int64  `tfsdk:"id"`
	UID      types.Int64  `tfsdk:"uid"`
	Username types.String `tfsdk:"username"`
	FullName types.String `tfsdk:"full_name"`
	Email    types.String `tfsdk:"email"`
	Group    types.Int64  `tfsdk:"group"`
	Home     types.String `tfsdk:"home"`
	Shell    types.String `tfsdk:"shell"`
	Locked   types.Bool   `tfsdk:"locked"`
	SMB      types.Bool   `tfsdk:"smb"`
	Sudo     types.Bool   `tfsdk:"sudo"`
	Builtin  types.Bool   `tfsdk:"builtin"`
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a user on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier for the user.",
				Optional:    true,
				Computed:    true,
			},
			"uid": schema.Int64Attribute{
				Description: "Unix user ID.",
				Optional:    true,
				Computed:    true,
			},
			"username": schema.StringAttribute{
				Description: "The username.",
				Optional:    true,
				Computed:    true,
			},
			"full_name": schema.StringAttribute{
				Description: "Full name of the user.",
				Computed:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email address.",
				Computed:    true,
			},
			"group": schema.Int64Attribute{
				Description: "Primary group ID.",
				Computed:    true,
			},
			"home": schema.StringAttribute{
				Description: "Home directory.",
				Computed:    true,
			},
			"shell": schema.StringAttribute{
				Description: "Login shell.",
				Computed:    true,
			},
			"locked": schema.BoolAttribute{
				Description: "Whether the account is locked.",
				Computed:    true,
			},
			"smb": schema.BoolAttribute{
				Description: "Whether SMB authentication is enabled.",
				Computed:    true,
			},
			"sudo": schema.BoolAttribute{
				Description: "Whether sudo access is enabled.",
				Computed:    true,
			},
			"builtin": schema.BoolAttribute{
				Description: "Whether this is a built-in user.",
				Computed:    true,
			},
		},
	}
}

func (d *UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config UserDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	var err error

	if !config.ID.IsNull() {
		err = d.client.GetInstance(ctx, "user", config.ID.ValueInt64(), &result)
	} else if !config.Username.IsNull() {
		params := client.NewQueryParams().WithFilter("username", "=", config.Username.ValueString())
		var results []map[string]interface{}
		err = d.client.Query(ctx, "user", params, &results)
		if err == nil && len(results) > 0 {
			result = results[0]
		} else if len(results) == 0 {
			resp.Diagnostics.AddError("User Not Found", fmt.Sprintf("User with username %s not found", config.Username.ValueString()))
			return
		}
	} else if !config.UID.IsNull() {
		params := client.NewQueryParams().WithFilter("uid", "=", config.UID.ValueInt64())
		var results []map[string]interface{}
		err = d.client.Query(ctx, "user", params, &results)
		if err == nil && len(results) > 0 {
			result = results[0]
		} else if len(results) == 0 {
			resp.Diagnostics.AddError("User Not Found", fmt.Sprintf("User with UID %d not found", config.UID.ValueInt64()))
			return
		}
	} else {
		resp.Diagnostics.AddError("Missing Identifier", "Either id, uid, or username must be specified")
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Error Reading User", "Could not read user: "+err.Error())
		return
	}

	config.ID = types.Int64Value(int64(result["id"].(float64)))
	config.UID = types.Int64Value(int64(result["uid"].(float64)))
	config.Username = types.StringValue(result["username"].(string))

	if fullName, ok := result["full_name"].(string); ok {
		config.FullName = types.StringValue(fullName)
	}
	if email, ok := result["email"].(string); ok {
		config.Email = types.StringValue(email)
	}
	if group, ok := result["group"].(map[string]interface{}); ok {
		if gid, ok := group["id"].(float64); ok {
			config.Group = types.Int64Value(int64(gid))
		}
	}
	if home, ok := result["home"].(string); ok {
		config.Home = types.StringValue(home)
	}
	if shell, ok := result["shell"].(string); ok {
		config.Shell = types.StringValue(shell)
	}
	if locked, ok := result["locked"].(bool); ok {
		config.Locked = types.BoolValue(locked)
	}
	if smb, ok := result["smb"].(bool); ok {
		config.SMB = types.BoolValue(smb)
	}
	if sudo, ok := result["sudo"].(bool); ok {
		config.Sudo = types.BoolValue(sudo)
	}
	if builtin, ok := result["builtin"].(bool); ok {
		config.Builtin = types.BoolValue(builtin)
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}
