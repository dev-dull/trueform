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
	_ resource.Resource                = &CertificateResource{}
	_ resource.ResourceWithImportState = &CertificateResource{}
)

func NewCertificateResource() resource.Resource {
	return &CertificateResource{}
}

type CertificateResource struct {
	client *client.Client
}

type CertificateResourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Type             types.String `tfsdk:"type"`
	Certificate      types.String `tfsdk:"certificate"`
	PrivateKey       types.String `tfsdk:"privatekey"`
	CSR              types.String `tfsdk:"csr"`
	SignedBy         types.Int64  `tfsdk:"signedby"`
	KeyLength        types.Int64  `tfsdk:"key_length"`
	KeyType          types.String `tfsdk:"key_type"`
	DigestAlgorithm  types.String `tfsdk:"digest_algorithm"`
	Lifetime         types.Int64  `tfsdk:"lifetime"`
	Country          types.String `tfsdk:"country"`
	State            types.String `tfsdk:"state"`
	City             types.String `tfsdk:"city"`
	Organization     types.String `tfsdk:"organization"`
	OrganizationalUnit types.String `tfsdk:"organizational_unit"`
	Email            types.String `tfsdk:"email"`
	CommonName       types.String `tfsdk:"common_name"`
	San              types.List   `tfsdk:"san"`
	CertificateChain types.Bool   `tfsdk:"cert_chain"`
	Fingerprint      types.String `tfsdk:"fingerprint"`
	NotBefore        types.String `tfsdk:"not_before"`
	NotAfter         types.String `tfsdk:"not_after"`
}

func (r *CertificateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (r *CertificateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a certificate on TrueNAS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier for the certificate.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the certificate.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Certificate creation type (CERTIFICATE_CREATE_INTERNAL, CERTIFICATE_CREATE_IMPORTED, CERTIFICATE_CREATE_CSR, CERTIFICATE_CREATE_ACME).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"certificate": schema.StringAttribute{
				Description: "PEM-encoded certificate (for imported certificates).",
				Optional:    true,
				Sensitive:   true,
			},
			"privatekey": schema.StringAttribute{
				Description: "PEM-encoded private key (for imported certificates).",
				Optional:    true,
				Sensitive:   true,
			},
			"csr": schema.StringAttribute{
				Description: "Certificate Signing Request.",
				Computed:    true,
			},
			"signedby": schema.Int64Attribute{
				Description: "ID of the CA that signed this certificate.",
				Optional:    true,
			},
			"key_length": schema.Int64Attribute{
				Description: "RSA key length (1024, 2048, 4096).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(2048),
			},
			"key_type": schema.StringAttribute{
				Description: "Key type (RSA, EC).",
				Optional:    true,
			},
			"digest_algorithm": schema.StringAttribute{
				Description: "Digest algorithm (SHA256, SHA384, SHA512).",
				Optional:    true,
			},
			"lifetime": schema.Int64Attribute{
				Description: "Certificate lifetime in days.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3650),
			},
			"country": schema.StringAttribute{
				Description: "Country code.",
				Optional:    true,
			},
			"state": schema.StringAttribute{
				Description: "State or province.",
				Optional:    true,
			},
			"city": schema.StringAttribute{
				Description: "City or locality.",
				Optional:    true,
			},
			"organization": schema.StringAttribute{
				Description: "Organization name.",
				Optional:    true,
			},
			"organizational_unit": schema.StringAttribute{
				Description: "Organizational unit.",
				Optional:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email address.",
				Optional:    true,
			},
			"common_name": schema.StringAttribute{
				Description: "Common name (CN).",
				Optional:    true,
			},
			"san": schema.ListAttribute{
				Description: "Subject Alternative Names.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"cert_chain": schema.BoolAttribute{
				Description: "Include certificate chain.",
				Optional:    true,
			},
			"fingerprint": schema.StringAttribute{
				Description: "Certificate fingerprint.",
				Computed:    true,
			},
			"not_before": schema.StringAttribute{
				Description: "Certificate validity start date.",
				Computed:    true,
			},
			"not_after": schema.StringAttribute{
				Description: "Certificate validity end date.",
				Computed:    true,
			},
		},
	}
}

func (r *CertificateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CertificateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating certificate", map[string]interface{}{
		"name": plan.Name.ValueString(),
		"type": plan.Type.ValueString(),
	})

	createData := map[string]interface{}{
		"name":       plan.Name.ValueString(),
		"create_type": plan.Type.ValueString(),
	}

	if !plan.Certificate.IsNull() {
		createData["certificate"] = plan.Certificate.ValueString()
	}
	if !plan.PrivateKey.IsNull() {
		createData["privatekey"] = plan.PrivateKey.ValueString()
	}
	if !plan.SignedBy.IsNull() {
		createData["signedby"] = plan.SignedBy.ValueInt64()
	}
	if !plan.KeyLength.IsNull() {
		createData["key_length"] = plan.KeyLength.ValueInt64()
	}
	if !plan.KeyType.IsNull() {
		createData["key_type"] = plan.KeyType.ValueString()
	}
	if !plan.DigestAlgorithm.IsNull() {
		createData["digest_algorithm"] = plan.DigestAlgorithm.ValueString()
	}
	if !plan.Lifetime.IsNull() {
		createData["lifetime"] = plan.Lifetime.ValueInt64()
	}
	if !plan.Country.IsNull() {
		createData["country"] = plan.Country.ValueString()
	}
	if !plan.State.IsNull() {
		createData["state"] = plan.State.ValueString()
	}
	if !plan.City.IsNull() {
		createData["city"] = plan.City.ValueString()
	}
	if !plan.Organization.IsNull() {
		createData["organization"] = plan.Organization.ValueString()
	}
	if !plan.OrganizationalUnit.IsNull() {
		createData["organizational_unit"] = plan.OrganizationalUnit.ValueString()
	}
	if !plan.Email.IsNull() {
		createData["email"] = plan.Email.ValueString()
	}
	if !plan.CommonName.IsNull() {
		createData["common_name"] = plan.CommonName.ValueString()
	}
	if !plan.San.IsNull() {
		var san []string
		diags = plan.San.ElementsAs(ctx, &san, false)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			createData["san"] = san
		}
	}

	var result map[string]interface{}
	err := r.client.Create(ctx, "certificate", createData, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Certificate", "Could not create certificate: "+err.Error())
		return
	}

	certID := int64(result["id"].(float64))
	if err := r.readCertificate(ctx, certID, &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading Certificate", "Could not read certificate after creation: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CertificateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readCertificate(ctx, state.ID.ValueInt64(), &state); err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Certificate", "Could not read certificate: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *CertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CertificateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CertificateResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Certificates have very limited update capability
	// Most fields require recreation
	updateData := map[string]interface{}{}

	if len(updateData) > 0 {
		var result map[string]interface{}
		err := r.client.Update(ctx, "certificate", state.ID.ValueInt64(), updateData, &result)
		if err != nil {
			resp.Diagnostics.AddError("Error Updating Certificate", "Could not update certificate: "+err.Error())
			return
		}
	}

	if err := r.readCertificate(ctx, state.ID.ValueInt64(), &plan); err != nil {
		resp.Diagnostics.AddError("Error Reading Certificate", "Could not read certificate after update: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CertificateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "certificate", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Certificate", "Could not delete certificate: "+err.Error())
		return
	}
}

func (r *CertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *CertificateResource) readCertificate(ctx context.Context, id int64, model *CertificateResourceModel) error {
	var result map[string]interface{}
	err := r.client.GetInstance(ctx, "certificate", id, &result)
	if err != nil {
		return err
	}

	model.ID = types.Int64Value(int64(result["id"].(float64)))
	model.Name = types.StringValue(result["name"].(string))

	if certType, ok := result["type"].(float64); ok {
		// Map numeric type back to string
		model.Type = types.StringValue(fmt.Sprintf("%d", int(certType)))
	}
	if csr, ok := result["CSR"].(string); ok {
		model.CSR = types.StringValue(csr)
	}
	if signedBy, ok := result["signedby"].(float64); ok {
		model.SignedBy = types.Int64Value(int64(signedBy))
	}
	if keyLength, ok := result["key_length"].(float64); ok {
		model.KeyLength = types.Int64Value(int64(keyLength))
	}
	if keyType, ok := result["key_type"].(string); ok {
		model.KeyType = types.StringValue(keyType)
	}
	if digestAlgorithm, ok := result["digest_algorithm"].(string); ok {
		model.DigestAlgorithm = types.StringValue(digestAlgorithm)
	}
	if lifetime, ok := result["lifetime"].(float64); ok {
		model.Lifetime = types.Int64Value(int64(lifetime))
	}
	if country, ok := result["country"].(string); ok {
		model.Country = types.StringValue(country)
	}
	if state, ok := result["state"].(string); ok {
		model.State = types.StringValue(state)
	}
	if city, ok := result["city"].(string); ok {
		model.City = types.StringValue(city)
	}
	if organization, ok := result["organization"].(string); ok {
		model.Organization = types.StringValue(organization)
	}
	if ou, ok := result["organizational_unit"].(string); ok {
		model.OrganizationalUnit = types.StringValue(ou)
	}
	if email, ok := result["email"].(string); ok {
		model.Email = types.StringValue(email)
	}
	if commonName, ok := result["common"].(string); ok {
		model.CommonName = types.StringValue(commonName)
	}
	if san, ok := result["san"].([]interface{}); ok {
		sanList := make([]string, len(san))
		for i, s := range san {
			sanList[i] = s.(string)
		}
		sanValues, diags := types.ListValueFrom(ctx, types.StringType, sanList)
		if !diags.HasError() {
			model.San = sanValues
		}
	}
	if fingerprint, ok := result["fingerprint"].(string); ok {
		model.Fingerprint = types.StringValue(fingerprint)
	}
	if notBefore, ok := result["from"].(string); ok {
		model.NotBefore = types.StringValue(notBefore)
	}
	if notAfter, ok := result["until"].(string); ok {
		model.NotAfter = types.StringValue(notAfter)
	}

	return nil
}
