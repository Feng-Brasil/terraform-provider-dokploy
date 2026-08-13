package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Feng-Brasil/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &DomainResource{}
var _ resource.ResourceWithImportState = &DomainResource{}

func NewDomainResource() resource.Resource {
	return &DomainResource{}
}

type DomainResource struct {
	client *client.DokployClient
}

type DomainResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	ApplicationID       types.String `tfsdk:"application_id"`
	ComposeID           types.String `tfsdk:"compose_id"`
	PreviewDeploymentID types.String `tfsdk:"preview_deployment_id"`
	ServiceName         types.String `tfsdk:"service_name"`
	Host                types.String `tfsdk:"host"`
	Path                types.String `tfsdk:"path"`
	Port                types.Int64  `tfsdk:"port"`
	HTTPS               types.Bool   `tfsdk:"https"`
	CertificateType     types.String `tfsdk:"certificate_type"`
	CustomEntrypoint    types.String `tfsdk:"custom_entrypoint"`
	CustomCertResolver  types.String `tfsdk:"custom_cert_resolver"`
	DomainType          types.String `tfsdk:"domain_type"`
	InternalPath        types.String `tfsdk:"internal_path"`
	StripPath           types.Bool   `tfsdk:"strip_path"`
	Middlewares         types.List   `tfsdk:"middlewares"`
	ForwardAuthEnabled  types.Bool   `tfsdk:"forward_auth_enabled"`
	GenerateTraefikMe   types.Bool   `tfsdk:"generate_traefik_me"`
	RedeployOnUpdate    types.Bool   `tfsdk:"redeploy_on_update"`
}

func (r *DomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (r *DomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"compose_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"preview_deployment_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Preview deployment ID for domains attached to preview deployments.",
			},
			"service_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"host": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"path": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"port": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"https": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable HTTPS for the domain.",
			},
			"certificate_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Certificate type: 'none', 'letsencrypt', or 'custom'. Defaults to 'letsencrypt' when https is true.",
			},
			"custom_entrypoint": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Custom Traefik entrypoint for this domain.",
			},
			"custom_cert_resolver": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Custom Traefik certificate resolver when certificate_type is 'custom'.",
			},
			"domain_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Domain type: application, compose, or preview.",
			},
			"internal_path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Internal path prefix for routing.",
			},
			"strip_path": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to strip the matched path prefix before forwarding.",
			},
			"middlewares": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of Traefik middleware names to apply.",
			},
			"forward_auth_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether forward authentication is enabled for the domain.",
			},
			"generate_traefik_me": schema.BoolAttribute{
				Optional:    true,
				Description: "If true, generates a traefik.me domain for the application.",
			},
			"redeploy_on_update": schema.BoolAttribute{
				Optional:    true,
				Description: "If true, triggers a redeploy of the associated application or compose stack when the domain is created or updated.",
			},
		},
	}
}

func (r *DomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Type", fmt.Sprintf("Expected *client.DokployClient, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *DomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DomainResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ApplicationID.IsNull() && plan.ComposeID.IsNull() && plan.PreviewDeploymentID.IsNull() {
		resp.Diagnostics.AddError("Missing Association", "One of application_id, compose_id, or preview_deployment_id must be provided")
		return
	}

	// Logic for domain generation
	if !plan.GenerateTraefikMe.IsNull() && plan.GenerateTraefikMe.ValueBool() {
		if plan.ApplicationID.IsNull() && plan.ComposeID.IsNull() {
			resp.Diagnostics.AddError(
				"Invalid Domain Generation Configuration",
				"generate_traefik_me requires application_id or compose_id to infer app name",
			)
			return
		}
		var name string
		if !plan.ApplicationID.IsNull() {
			app, err := r.client.GetApplication(plan.ApplicationID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Error fetching application for domain generation", err.Error())
				return
			}
			name = app.Name
		} else {
			comp, err := r.client.GetCompose(plan.ComposeID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Error fetching compose for domain generation", err.Error())
				return
			}
			name = comp.Name
		}

		generatedDomain, err := r.client.GenerateDomain(name)
		if err != nil {
			resp.Diagnostics.AddError("Error generating traefik.me domain", err.Error())
			return
		}
		plan.Host = types.StringValue(generatedDomain)
	} else {
		if plan.Host.IsNull() || plan.Host.IsUnknown() {
			resp.Diagnostics.AddError("Missing Host", "Host is required when generate_traefik_me is false")
			return
		}
	}

	// Apply defaults
	if plan.Path.IsUnknown() || plan.Path.IsNull() {
		plan.Path = types.StringValue("/")
	}
	if plan.Port.IsUnknown() || plan.Port.IsNull() {
		plan.Port = types.Int64Value(3000)
	}
	if plan.HTTPS.IsUnknown() || plan.HTTPS.IsNull() {
		plan.HTTPS = types.BoolValue(true)
	}
	if plan.StripPath.IsUnknown() || plan.StripPath.IsNull() {
		plan.StripPath = types.BoolValue(false)
	}
	if plan.ForwardAuthEnabled.IsUnknown() || plan.ForwardAuthEnabled.IsNull() {
		plan.ForwardAuthEnabled = types.BoolValue(false)
	}

	var middlewares []string
	if !plan.Middlewares.IsNull() && !plan.Middlewares.IsUnknown() {
		diags = plan.Middlewares.ElementsAs(ctx, &middlewares, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	domain := client.Domain{
		ApplicationID:       plan.ApplicationID.ValueString(),
		ComposeID:           plan.ComposeID.ValueString(),
		PreviewDeploymentID: plan.PreviewDeploymentID.ValueString(),
		ServiceName:         plan.ServiceName.ValueString(),
		Host:                plan.Host.ValueString(),
		Path:                plan.Path.ValueString(),
		Port:                plan.Port.ValueInt64(),
		HTTPS:               plan.HTTPS.ValueBool(),
		CertificateType:     plan.CertificateType.ValueString(),
		CustomEntrypoint:    plan.CustomEntrypoint.ValueString(),
		CustomCertResolver:  plan.CustomCertResolver.ValueString(),
		DomainType:          plan.DomainType.ValueString(),
		InternalPath:        plan.InternalPath.ValueString(),
		StripPath:           plan.StripPath.ValueBool(),
		Middlewares:         middlewares,
		ForwardAuthEnabled:  plan.ForwardAuthEnabled.ValueBool(),
	}

	createdDomain, err := r.client.CreateDomain(domain)
	if err != nil {
		resp.Diagnostics.AddError("Error creating domain", err.Error())
		return
	}

	if createdDomain.ID == "" {
		resp.Diagnostics.AddError("Error creating domain", "API response did not include domainId")
		return
	}

	plan.ID = types.StringValue(createdDomain.ID)
	fullDomain, err := r.client.GetDomain(createdDomain.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading created domain", err.Error())
		return
	}
	plan = domainModelFromClient(ctx, plan, fullDomain, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Trigger Redeploy if requested
	if !plan.RedeployOnUpdate.IsNull() && plan.RedeployOnUpdate.ValueBool() {
		if !plan.ApplicationID.IsNull() {
			_ = r.client.DeployApplication(plan.ApplicationID.ValueString(), "")
		} else if !plan.ComposeID.IsNull() {
			_ = r.client.DeployCompose(plan.ComposeID.ValueString(), "")
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DomainResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.client.GetDomain(state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) || strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading domain", err.Error())
		return
	}

	state = domainModelFromClient(ctx, state, domain, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *DomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DomainResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var middlewares []string
	if !plan.Middlewares.IsNull() && !plan.Middlewares.IsUnknown() {
		diags = plan.Middlewares.ElementsAs(ctx, &middlewares, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	domain := client.Domain{
		ID:                  plan.ID.ValueString(),
		ApplicationID:       plan.ApplicationID.ValueString(),
		ComposeID:           plan.ComposeID.ValueString(),
		PreviewDeploymentID: plan.PreviewDeploymentID.ValueString(),
		ServiceName:         plan.ServiceName.ValueString(),
		Host:                plan.Host.ValueString(),
		Path:                plan.Path.ValueString(),
		Port:                plan.Port.ValueInt64(),
		HTTPS:               plan.HTTPS.ValueBool(),
		CertificateType:     plan.CertificateType.ValueString(),
		CustomEntrypoint:    plan.CustomEntrypoint.ValueString(),
		CustomCertResolver:  plan.CustomCertResolver.ValueString(),
		DomainType:          plan.DomainType.ValueString(),
		InternalPath:        plan.InternalPath.ValueString(),
		StripPath:           plan.StripPath.ValueBool(),
		Middlewares:         middlewares,
		ForwardAuthEnabled:  plan.ForwardAuthEnabled.ValueBool(),
	}

	updatedDomain, err := r.client.UpdateDomain(domain)
	if err != nil {
		resp.Diagnostics.AddError("Error updating domain", err.Error())
		return
	}

	domainID := plan.ID.ValueString()
	if updatedDomain.ID != "" {
		domainID = updatedDomain.ID
	}

	fullDomain, err := r.client.GetDomain(domainID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading updated domain", err.Error())
		return
	}
	plan = domainModelFromClient(ctx, plan, fullDomain, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Trigger Redeploy if requested
	if !plan.RedeployOnUpdate.IsNull() && plan.RedeployOnUpdate.ValueBool() {
		if !plan.ApplicationID.IsNull() {
			_ = r.client.DeployApplication(plan.ApplicationID.ValueString(), "")
		} else if !plan.ComposeID.IsNull() {
			_ = r.client.DeployCompose(plan.ComposeID.ValueString(), "")
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DomainResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDomain(state.ID.ValueString())
	if err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "not found") || strings.Contains(errStr, "not_found") || strings.Contains(errStr, "404") {
			// Resource already deleted, that's fine
			return
		}
		resp.Diagnostics.AddError("Error deleting domain", err.Error())
		return
	}
}

func (r *DomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importID := req.ID

	// Try to parse the import ID
	var parentType, parentID, domainID string

	parts := strings.Split(importID, ":")
	if len(parts) == 3 {
		// Format: application:app-id:domain-id, compose:compose-id:domain-id, or preview:preview-deployment-id:domain-id
		parentType = parts[0]
		parentID = parts[1]
		domainID = parts[2]
	} else if len(parts) == 1 {
		// Format: <domain-id>
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), importID)...)
		return
	} else {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			fmt.Sprintf("Expected '<domain-id>' or '<type>:<parent-id>:<domain-id>' with type in [application|compose|preview]. Got: %s", importID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), domainID)...)

	switch parentType {
	case "application":
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), parentID)...)
	case "compose":
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("compose_id"), parentID)...)
	case "preview":
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("preview_deployment_id"), parentID)...)
	default:
		resp.Diagnostics.AddError(
			"Invalid parent type",
			fmt.Sprintf("Parent type must be 'application', 'compose', or 'preview'. Got: %s", parentType),
		)
		return
	}
}

func domainModelFromClient(ctx context.Context, model DomainResourceModel, d *client.Domain, diags *diag.Diagnostics) DomainResourceModel {
	model.ID = types.StringValue(d.ID)
	model.Host = types.StringValue(d.Host)
	model.Path = types.StringValue(d.Path)
	model.Port = types.Int64Value(d.Port)
	model.HTTPS = types.BoolValue(d.HTTPS)
	model.StripPath = types.BoolValue(d.StripPath)
	model.ForwardAuthEnabled = types.BoolValue(d.ForwardAuthEnabled)

	if d.ApplicationID != "" {
		model.ApplicationID = types.StringValue(d.ApplicationID)
	} else {
		model.ApplicationID = types.StringNull()
	}
	if d.ComposeID != "" {
		model.ComposeID = types.StringValue(d.ComposeID)
	} else {
		model.ComposeID = types.StringNull()
	}
	if d.PreviewDeploymentID != "" {
		model.PreviewDeploymentID = types.StringValue(d.PreviewDeploymentID)
	} else {
		model.PreviewDeploymentID = types.StringNull()
	}
	if d.ServiceName != "" {
		model.ServiceName = types.StringValue(d.ServiceName)
	} else {
		model.ServiceName = types.StringNull()
	}
	if d.CertificateType != "" {
		model.CertificateType = types.StringValue(d.CertificateType)
	} else {
		model.CertificateType = types.StringNull()
	}
	if d.CustomEntrypoint != "" {
		model.CustomEntrypoint = types.StringValue(d.CustomEntrypoint)
	} else {
		model.CustomEntrypoint = types.StringNull()
	}
	if d.CustomCertResolver != "" {
		model.CustomCertResolver = types.StringValue(d.CustomCertResolver)
	} else {
		model.CustomCertResolver = types.StringNull()
	}
	if d.DomainType != "" {
		model.DomainType = types.StringValue(d.DomainType)
	} else {
		model.DomainType = types.StringNull()
	}
	if d.InternalPath != "" {
		model.InternalPath = types.StringValue(d.InternalPath)
	} else {
		model.InternalPath = types.StringNull()
	}
	if d.Middlewares == nil {
		model.Middlewares = types.ListNull(types.StringType)
	} else {
		listVal, listDiags := types.ListValueFrom(ctx, types.StringType, d.Middlewares)
		diags.Append(listDiags...)
		model.Middlewares = listVal
	}

	return model
}
