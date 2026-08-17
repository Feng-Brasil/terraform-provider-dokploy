package provider

import (
	"context"
	"fmt"

	"github.com/feng-brasil/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &UsersDataSource{}

func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

type UsersDataSource struct {
	client *client.DokployClient
}

type UsersDataSourceModel struct {
	Users []UserModel `tfsdk:"users"`
}

type UserModel struct {
	// Member fields
	MemberID       types.String `tfsdk:"member_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	UserID         types.String `tfsdk:"user_id"`
	Role           types.String `tfsdk:"role"`
	TeamID         types.String `tfsdk:"team_id"`
	IsDefault      types.Bool   `tfsdk:"is_default"`
	CreatedAt      types.String `tfsdk:"created_at"`

	// Permission fields
	CanCreateProjects       types.Bool `tfsdk:"can_create_projects"`
	CanAccessToSSHKeys      types.Bool `tfsdk:"can_access_to_ssh_keys"`
	CanCreateServices       types.Bool `tfsdk:"can_create_services"`
	CanDeleteProjects       types.Bool `tfsdk:"can_delete_projects"`
	CanDeleteServices       types.Bool `tfsdk:"can_delete_services"`
	CanAccessToDocker       types.Bool `tfsdk:"can_access_to_docker"`
	CanAccessToAPI          types.Bool `tfsdk:"can_access_to_api"`
	CanAccessToGitProviders types.Bool `tfsdk:"can_access_to_git_providers"`
	CanAccessToTraefikFiles types.Bool `tfsdk:"can_access_to_traefik_files"`
	CanDeleteEnvironments   types.Bool `tfsdk:"can_delete_environments"`
	CanCreateEnvironments   types.Bool `tfsdk:"can_create_environments"`
	AccessedProjects        types.List `tfsdk:"accessed_projects"`
	AccessedEnvironments    types.List `tfsdk:"accessed_environments"`
	AccessedServices        types.List `tfsdk:"accessed_services"`

	// User details
	FirstName        types.String `tfsdk:"first_name"`
	LastName         types.String `tfsdk:"last_name"`
	Email            types.String `tfsdk:"email"`
	EmailVerified    types.Bool   `tfsdk:"email_verified"`
	TwoFactorEnabled types.Bool   `tfsdk:"two_factor_enabled"`
	Image            types.String `tfsdk:"image"`
}

func (d *UsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all users (organization members) in the current Dokploy organization.",
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of users in the organization.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"member_id": schema.StringAttribute{
							Computed:    true,
							Description: "The organization membership ID.",
						},
						"organization_id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the organization.",
						},
						"user_id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique user ID.",
						},
						"role": schema.StringAttribute{
							Computed:    true,
							Description: "The user's role in the organization (e.g., 'owner', 'member').",
						},
						"team_id": schema.StringAttribute{
							Computed:    true,
							Description: "The team ID if the user belongs to a team.",
						},
						"is_default": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether this is the default organization membership.",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "The timestamp when the membership was created.",
						},

						// Permission fields
						"can_create_projects": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the user can create projects.",
						},
						"can_access_to_ssh_keys": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the user can access SSH keys.",
						},
						"can_create_services": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the user can create services.",
						},
						"can_delete_projects": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the user can delete projects.",
						},
						"can_delete_services": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the user can delete services.",
						},
						"can_access_to_docker": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the user can access Docker.",
						},
						"can_access_to_api": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the user can access the API.",
						},
						"can_access_to_git_providers": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the user can access Git providers.",
						},
						"can_access_to_traefik_files": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the user can access Traefik files.",
						},
						"can_delete_environments": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the user can delete environments.",
						},
						"can_create_environments": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the user can create environments.",
						},
						"accessed_projects": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "List of project IDs the user has access to.",
						},
						"accessed_environments": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "List of environment IDs the user has access to.",
						},
						"accessed_services": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "List of service IDs the user has access to.",
						},

						// User details
						"first_name": schema.StringAttribute{
							Computed:    true,
							Description: "The user's first name.",
						},
						"last_name": schema.StringAttribute{
							Computed:    true,
							Description: "The user's last name.",
						},
						"email": schema.StringAttribute{
							Computed:    true,
							Description: "The user's email address.",
						},
						"email_verified": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the user's email is verified.",
						},
						"two_factor_enabled": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether two-factor authentication is enabled.",
						},
						"image": schema.StringAttribute{
							Computed:    true,
							Description: "The user's profile image URL.",
						},
					},
				},
			},
		},
	}
}

func (d *UsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *client.DokployClient, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	members, err := d.client.ListMembers()
	if err != nil {
		resp.Diagnostics.AddError("Unable to List Users", err.Error())
		return
	}

	var state UsersDataSourceModel

	for _, member := range members {
		userModel := UserModel{
			MemberID:       types.StringValue(member.ID),
			OrganizationID: types.StringValue(member.OrganizationID),
			UserID:         types.StringValue(member.UserID),
			Role:           types.StringValue(member.Role),
			IsDefault:      types.BoolValue(member.IsDefault),
			CreatedAt:      types.StringValue(member.CreatedAt),

			// Permissions
			CanCreateProjects:       types.BoolValue(member.CanCreateProjects),
			CanAccessToSSHKeys:      types.BoolValue(member.CanAccessToSSHKeys),
			CanCreateServices:       types.BoolValue(member.CanCreateServices),
			CanDeleteProjects:       types.BoolValue(member.CanDeleteProjects),
			CanDeleteServices:       types.BoolValue(member.CanDeleteServices),
			CanAccessToDocker:       types.BoolValue(member.CanAccessToDocker),
			CanAccessToAPI:          types.BoolValue(member.CanAccessToAPI),
			CanAccessToGitProviders: types.BoolValue(member.CanAccessToGitProviders),
			CanAccessToTraefikFiles: types.BoolValue(member.CanAccessToTraefikFiles),
			CanDeleteEnvironments:   types.BoolValue(member.CanDeleteEnvironments),
			CanCreateEnvironments:   types.BoolValue(member.CanCreateEnvironments),

			// User details
			FirstName:        types.StringValue(member.User.FirstName),
			LastName:         types.StringValue(member.User.LastName),
			Email:            types.StringValue(member.User.Email),
			EmailVerified:    types.BoolValue(member.User.EmailVerified),
			TwoFactorEnabled: types.BoolValue(member.User.TwoFactorEnabled),
		}

		// Handle optional team_id
		if member.TeamID != nil {
			userModel.TeamID = types.StringValue(*member.TeamID)
		} else {
			userModel.TeamID = types.StringNull()
		}

		// Handle optional image
		if member.User.Image != nil {
			userModel.Image = types.StringValue(*member.User.Image)
		} else {
			userModel.Image = types.StringNull()
		}

		// Convert string slices to list values
		accessedProjects, diags := types.ListValueFrom(ctx, types.StringType, member.AccessedProjects)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		userModel.AccessedProjects = accessedProjects

		accessedEnvironments, diags := types.ListValueFrom(ctx, types.StringType, member.AccessedEnvironments)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		userModel.AccessedEnvironments = accessedEnvironments

		accessedServices, diags := types.ListValueFrom(ctx, types.StringType, member.AccessedServices)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		userModel.AccessedServices = accessedServices

		state.Users = append(state.Users, userModel)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
