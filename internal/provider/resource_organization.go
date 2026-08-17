package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/feng-brasil/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &OrganizationResource{}
var _ resource.ResourceWithImportState = &OrganizationResource{}

func NewOrganizationResource() resource.Resource {
	return &OrganizationResource{}
}

type OrganizationResource struct {
	client *client.DokployClient
}

type OrganizationResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Logo      types.String `tfsdk:"logo"`
	Slug      types.String `tfsdk:"slug"`
	OwnerID   types.String `tfsdk:"owner_id"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (r *OrganizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (r *OrganizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an organization in Dokploy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the organization.",
			},
			"logo": schema.StringAttribute{
				Optional:    true,
				Description: "URL or path to the organization logo.",
			},
			"slug": schema.StringAttribute{
				Computed:    true,
				Description: "URL-friendly identifier for the organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"owner_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the user who owns the organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the organization was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *OrganizationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.DokployClient, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *OrganizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OrganizationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var logo *string
	if !plan.Logo.IsNull() && !plan.Logo.IsUnknown() {
		logoVal := plan.Logo.ValueString()
		logo = &logoVal
	}

	org, err := r.client.CreateOrganization(plan.Name.ValueString(), logo)
	if err != nil {
		resp.Diagnostics.AddError("Error creating organization", err.Error())
		return
	}

	plan.ID = types.StringValue(org.ID)
	plan.Name = types.StringValue(org.Name)
	plan.OwnerID = types.StringValue(org.OwnerID)
	plan.CreatedAt = types.StringValue(org.CreatedAt)

	if org.Slug != nil {
		plan.Slug = types.StringValue(*org.Slug)
	} else {
		plan.Slug = types.StringNull()
	}

	if org.Logo != nil {
		plan.Logo = types.StringValue(*org.Logo)
	} else if plan.Logo.IsUnknown() {
		plan.Logo = types.StringNull()
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *OrganizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OrganizationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.client.GetOrganization(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading organization", err.Error())
		return
	}

	state.Name = types.StringValue(org.Name)
	state.OwnerID = types.StringValue(org.OwnerID)
	state.CreatedAt = types.StringValue(org.CreatedAt)

	if org.Slug != nil {
		state.Slug = types.StringValue(*org.Slug)
	} else {
		state.Slug = types.StringNull()
	}

	if org.Logo != nil {
		state.Logo = types.StringValue(*org.Logo)
	} else {
		state.Logo = types.StringNull()
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *OrganizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OrganizationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state OrganizationResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgUpdate := client.Organization{
		ID:   state.ID.ValueString(),
		Name: plan.Name.ValueString(),
	}

	if !plan.Logo.IsNull() && !plan.Logo.IsUnknown() {
		logoVal := plan.Logo.ValueString()
		orgUpdate.Logo = &logoVal
	}

	org, err := r.client.UpdateOrganization(orgUpdate)
	if err != nil {
		resp.Diagnostics.AddError("Error updating organization", err.Error())
		return
	}

	plan.ID = types.StringValue(org.ID)
	plan.Name = types.StringValue(org.Name)
	plan.OwnerID = types.StringValue(org.OwnerID)
	plan.CreatedAt = types.StringValue(org.CreatedAt)

	if org.Slug != nil {
		plan.Slug = types.StringValue(*org.Slug)
	} else {
		plan.Slug = types.StringNull()
	}

	if org.Logo != nil {
		plan.Logo = types.StringValue(*org.Logo)
	} else if plan.Logo.IsNull() {
		plan.Logo = types.StringNull()
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *OrganizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OrganizationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteOrganization(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError("Error deleting organization", err.Error())
		return
	}
}

func (r *OrganizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
