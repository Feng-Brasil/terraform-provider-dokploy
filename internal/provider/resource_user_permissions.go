package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/Feng-Brasil/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &UserPermissionsResource{}
var _ resource.ResourceWithImportState = &UserPermissionsResource{}

func NewUserPermissionsResource() resource.Resource {
	return &UserPermissionsResource{}
}

type UserPermissionsResource struct {
	client *client.DokployClient
}

type UserPermissionsResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	MemberID                types.String `tfsdk:"member_id"`
	CanCreateProjects       types.Bool   `tfsdk:"can_create_projects"`
	CanAccessToSSHKeys      types.Bool   `tfsdk:"can_access_to_ssh_keys"`
	CanCreateServices       types.Bool   `tfsdk:"can_create_services"`
	CanDeleteProjects       types.Bool   `tfsdk:"can_delete_projects"`
	CanDeleteServices       types.Bool   `tfsdk:"can_delete_services"`
	CanAccessToDocker       types.Bool   `tfsdk:"can_access_to_docker"`
	CanAccessToAPI          types.Bool   `tfsdk:"can_access_to_api"`
	CanAccessToGitProviders types.Bool   `tfsdk:"can_access_to_git_providers"`
	CanAccessToTraefikFiles types.Bool   `tfsdk:"can_access_to_traefik_files"`
	CanDeleteEnvironments   types.Bool   `tfsdk:"can_delete_environments"`
	CanCreateEnvironments   types.Bool   `tfsdk:"can_create_environments"`
	AccessedProjects        types.List   `tfsdk:"accessed_projects"`
	AccessedEnvironments    types.List   `tfsdk:"accessed_environments"`
	AccessedServices        types.List   `tfsdk:"accessed_services"`
}

func (r *UserPermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_permissions"
}

func (r *UserPermissionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages user permissions for an organization member in Dokploy. Note: Owner permissions cannot be modified.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier (same as member_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"member_id": schema.StringAttribute{
				Required:    true,
				Description: "The organization membership ID of the user whose permissions to manage. Use the 'member_id' from dokploy_user or dokploy_users data sources.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"can_create_projects": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the user can create projects. Defaults to false.",
			},
			"can_access_to_ssh_keys": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the user can access SSH keys. Defaults to false.",
			},
			"can_create_services": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the user can create services. Defaults to false.",
			},
			"can_delete_projects": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the user can delete projects. Defaults to false.",
			},
			"can_delete_services": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the user can delete services. Defaults to false.",
			},
			"can_access_to_docker": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the user can access Docker. Defaults to false.",
			},
			"can_access_to_api": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the user can access the API. Defaults to false.",
			},
			"can_access_to_git_providers": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the user can access Git providers. Defaults to false.",
			},
			"can_access_to_traefik_files": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the user can access Traefik files. Defaults to false.",
			},
			"can_delete_environments": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the user can delete environments. Defaults to false.",
			},
			"can_create_environments": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the user can create environments. Defaults to false.",
			},
			"accessed_projects": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of project IDs the user has access to. Defaults to empty list.",
			},
			"accessed_environments": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of environment IDs the user has access to. Defaults to empty list.",
			},
			"accessed_services": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of service IDs the user has access to. Defaults to empty list.",
			},
		},
	}
}

func (r *UserPermissionsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserPermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserPermissionsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the accessed lists
	var accessedProjects, accessedEnvironments, accessedServices []string

	if !plan.AccessedProjects.IsNull() && !plan.AccessedProjects.IsUnknown() {
		diags = plan.AccessedProjects.ElementsAs(ctx, &accessedProjects, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		accessedProjects = []string{}
	}

	if !plan.AccessedEnvironments.IsNull() && !plan.AccessedEnvironments.IsUnknown() {
		diags = plan.AccessedEnvironments.ElementsAs(ctx, &accessedEnvironments, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		accessedEnvironments = []string{}
	}

	if !plan.AccessedServices.IsNull() && !plan.AccessedServices.IsUnknown() {
		diags = plan.AccessedServices.ElementsAs(ctx, &accessedServices, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		accessedServices = []string{}
	}

	input := client.UserPermissionsInput{
		MemberID:                plan.MemberID.ValueString(),
		AccessedProjects:        accessedProjects,
		AccessedEnvironments:    accessedEnvironments,
		AccessedServices:        accessedServices,
		CanCreateProjects:       plan.CanCreateProjects.ValueBool(),
		CanCreateServices:       plan.CanCreateServices.ValueBool(),
		CanDeleteProjects:       plan.CanDeleteProjects.ValueBool(),
		CanDeleteServices:       plan.CanDeleteServices.ValueBool(),
		CanAccessToDocker:       plan.CanAccessToDocker.ValueBool(),
		CanAccessToTraefikFiles: plan.CanAccessToTraefikFiles.ValueBool(),
		CanAccessToAPI:          plan.CanAccessToAPI.ValueBool(),
		CanAccessToSSHKeys:      plan.CanAccessToSSHKeys.ValueBool(),
		CanAccessToGitProviders: plan.CanAccessToGitProviders.ValueBool(),
		CanDeleteEnvironments:   plan.CanDeleteEnvironments.ValueBool(),
		CanCreateEnvironments:   plan.CanCreateEnvironments.ValueBool(),
	}

	err := r.client.AssignUserPermissions(input)
	if err != nil {
		resp.Diagnostics.AddError("Error assigning user permissions", err.Error())
		return
	}

	// Read back the member to get current state
	member, err := r.client.GetMemberByID(plan.MemberID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading user after permission assignment", err.Error())
		return
	}

	plan.ID = types.StringValue(member.ID)
	plan.CanCreateProjects = types.BoolValue(member.CanCreateProjects)
	plan.CanAccessToSSHKeys = types.BoolValue(member.CanAccessToSSHKeys)
	plan.CanCreateServices = types.BoolValue(member.CanCreateServices)
	plan.CanDeleteProjects = types.BoolValue(member.CanDeleteProjects)
	plan.CanDeleteServices = types.BoolValue(member.CanDeleteServices)
	plan.CanAccessToDocker = types.BoolValue(member.CanAccessToDocker)
	plan.CanAccessToAPI = types.BoolValue(member.CanAccessToAPI)
	plan.CanAccessToGitProviders = types.BoolValue(member.CanAccessToGitProviders)
	plan.CanAccessToTraefikFiles = types.BoolValue(member.CanAccessToTraefikFiles)
	plan.CanDeleteEnvironments = types.BoolValue(member.CanDeleteEnvironments)
	plan.CanCreateEnvironments = types.BoolValue(member.CanCreateEnvironments)

	// Convert lists
	accessedProjectsList, diags := types.ListValueFrom(ctx, types.StringType, member.AccessedProjects)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.AccessedProjects = accessedProjectsList

	accessedEnvironmentsList, diags := types.ListValueFrom(ctx, types.StringType, member.AccessedEnvironments)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.AccessedEnvironments = accessedEnvironmentsList

	accessedServicesList, diags := types.ListValueFrom(ctx, types.StringType, member.AccessedServices)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.AccessedServices = accessedServicesList

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *UserPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserPermissionsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	member, err := r.client.GetMemberByID(state.MemberID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading user permissions", err.Error())
		return
	}

	state.ID = types.StringValue(member.ID)
	state.CanCreateProjects = types.BoolValue(member.CanCreateProjects)
	state.CanAccessToSSHKeys = types.BoolValue(member.CanAccessToSSHKeys)
	state.CanCreateServices = types.BoolValue(member.CanCreateServices)
	state.CanDeleteProjects = types.BoolValue(member.CanDeleteProjects)
	state.CanDeleteServices = types.BoolValue(member.CanDeleteServices)
	state.CanAccessToDocker = types.BoolValue(member.CanAccessToDocker)
	state.CanAccessToAPI = types.BoolValue(member.CanAccessToAPI)
	state.CanAccessToGitProviders = types.BoolValue(member.CanAccessToGitProviders)
	state.CanAccessToTraefikFiles = types.BoolValue(member.CanAccessToTraefikFiles)
	state.CanDeleteEnvironments = types.BoolValue(member.CanDeleteEnvironments)
	state.CanCreateEnvironments = types.BoolValue(member.CanCreateEnvironments)

	// Convert lists
	accessedProjectsList, diags := types.ListValueFrom(ctx, types.StringType, member.AccessedProjects)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.AccessedProjects = accessedProjectsList

	accessedEnvironmentsList, diags := types.ListValueFrom(ctx, types.StringType, member.AccessedEnvironments)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.AccessedEnvironments = accessedEnvironmentsList

	accessedServicesList, diags := types.ListValueFrom(ctx, types.StringType, member.AccessedServices)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.AccessedServices = accessedServicesList

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *UserPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserPermissionsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the accessed lists
	var accessedProjects, accessedEnvironments, accessedServices []string

	if !plan.AccessedProjects.IsNull() && !plan.AccessedProjects.IsUnknown() {
		diags = plan.AccessedProjects.ElementsAs(ctx, &accessedProjects, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		accessedProjects = []string{}
	}

	if !plan.AccessedEnvironments.IsNull() && !plan.AccessedEnvironments.IsUnknown() {
		diags = plan.AccessedEnvironments.ElementsAs(ctx, &accessedEnvironments, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		accessedEnvironments = []string{}
	}

	if !plan.AccessedServices.IsNull() && !plan.AccessedServices.IsUnknown() {
		diags = plan.AccessedServices.ElementsAs(ctx, &accessedServices, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		accessedServices = []string{}
	}

	input := client.UserPermissionsInput{
		MemberID:                plan.MemberID.ValueString(),
		AccessedProjects:        accessedProjects,
		AccessedEnvironments:    accessedEnvironments,
		AccessedServices:        accessedServices,
		CanCreateProjects:       plan.CanCreateProjects.ValueBool(),
		CanCreateServices:       plan.CanCreateServices.ValueBool(),
		CanDeleteProjects:       plan.CanDeleteProjects.ValueBool(),
		CanDeleteServices:       plan.CanDeleteServices.ValueBool(),
		CanAccessToDocker:       plan.CanAccessToDocker.ValueBool(),
		CanAccessToTraefikFiles: plan.CanAccessToTraefikFiles.ValueBool(),
		CanAccessToAPI:          plan.CanAccessToAPI.ValueBool(),
		CanAccessToSSHKeys:      plan.CanAccessToSSHKeys.ValueBool(),
		CanAccessToGitProviders: plan.CanAccessToGitProviders.ValueBool(),
		CanDeleteEnvironments:   plan.CanDeleteEnvironments.ValueBool(),
		CanCreateEnvironments:   plan.CanCreateEnvironments.ValueBool(),
	}

	err := r.client.AssignUserPermissions(input)
	if err != nil {
		resp.Diagnostics.AddError("Error updating user permissions", err.Error())
		return
	}

	// Read back the member to get current state
	member, err := r.client.GetMemberByID(plan.MemberID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading user after permission update", err.Error())
		return
	}

	plan.ID = types.StringValue(member.ID)
	plan.CanCreateProjects = types.BoolValue(member.CanCreateProjects)
	plan.CanAccessToSSHKeys = types.BoolValue(member.CanAccessToSSHKeys)
	plan.CanCreateServices = types.BoolValue(member.CanCreateServices)
	plan.CanDeleteProjects = types.BoolValue(member.CanDeleteProjects)
	plan.CanDeleteServices = types.BoolValue(member.CanDeleteServices)
	plan.CanAccessToDocker = types.BoolValue(member.CanAccessToDocker)
	plan.CanAccessToAPI = types.BoolValue(member.CanAccessToAPI)
	plan.CanAccessToGitProviders = types.BoolValue(member.CanAccessToGitProviders)
	plan.CanAccessToTraefikFiles = types.BoolValue(member.CanAccessToTraefikFiles)
	plan.CanDeleteEnvironments = types.BoolValue(member.CanDeleteEnvironments)
	plan.CanCreateEnvironments = types.BoolValue(member.CanCreateEnvironments)

	// Convert lists
	accessedProjectsList, diags := types.ListValueFrom(ctx, types.StringType, member.AccessedProjects)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.AccessedProjects = accessedProjectsList

	accessedEnvironmentsList, diags := types.ListValueFrom(ctx, types.StringType, member.AccessedEnvironments)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.AccessedEnvironments = accessedEnvironmentsList

	accessedServicesList, diags := types.ListValueFrom(ctx, types.StringType, member.AccessedServices)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.AccessedServices = accessedServicesList

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *UserPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserPermissionsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// On delete, reset all permissions to false
	input := client.UserPermissionsInput{
		MemberID:                state.MemberID.ValueString(),
		AccessedProjects:        []string{},
		AccessedEnvironments:    []string{},
		AccessedServices:        []string{},
		CanCreateProjects:       false,
		CanCreateServices:       false,
		CanDeleteProjects:       false,
		CanDeleteServices:       false,
		CanAccessToDocker:       false,
		CanAccessToTraefikFiles: false,
		CanAccessToAPI:          false,
		CanAccessToSSHKeys:      false,
		CanAccessToGitProviders: false,
		CanDeleteEnvironments:   false,
		CanCreateEnvironments:   false,
	}

	err := r.client.AssignUserPermissions(input)
	if err != nil {
		// If user is not found, consider it deleted
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError("Error resetting user permissions", err.Error())
		return
	}
}

func (r *UserPermissionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import using member_id
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("member_id"), req.ID)...)
}
