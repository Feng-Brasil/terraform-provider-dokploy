package provider

import (
	"context"
	"fmt"

	"github.com/feng-brasil/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &OrganizationsDataSource{}

func NewOrganizationsDataSource() datasource.DataSource {
	return &OrganizationsDataSource{}
}

type OrganizationsDataSource struct {
	client *client.DokployClient
}

type OrganizationsDataSourceModel struct {
	Organizations []OrganizationModel `tfsdk:"organizations"`
}

type OrganizationModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Slug      types.String `tfsdk:"slug"`
	Logo      types.String `tfsdk:"logo"`
	OwnerID   types.String `tfsdk:"owner_id"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (d *OrganizationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organizations"
}

func (d *OrganizationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of organizations in Dokploy.",
		Attributes: map[string]schema.Attribute{
			"organizations": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of organizations.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the organization.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the organization.",
						},
						"slug": schema.StringAttribute{
							Computed:    true,
							Description: "The URL-friendly identifier of the organization.",
						},
						"logo": schema.StringAttribute{
							Computed:    true,
							Description: "The logo URL of the organization.",
						},
						"owner_id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the user who owns the organization.",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "The creation timestamp of the organization.",
						},
					},
				},
			},
		},
	}
}

func (d *OrganizationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrganizationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config OrganizationsDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgs, err := d.client.ListOrganizations()
	if err != nil {
		resp.Diagnostics.AddError("Unable to List Organizations", err.Error())
		return
	}

	var state OrganizationsDataSourceModel

	for _, org := range orgs {
		orgModel := OrganizationModel{
			ID:        types.StringValue(org.ID),
			Name:      types.StringValue(org.Name),
			OwnerID:   types.StringValue(org.OwnerID),
			CreatedAt: types.StringValue(org.CreatedAt),
		}

		if org.Slug != nil {
			orgModel.Slug = types.StringValue(*org.Slug)
		} else {
			orgModel.Slug = types.StringNull()
		}

		if org.Logo != nil {
			orgModel.Logo = types.StringValue(*org.Logo)
		} else {
			orgModel.Logo = types.StringNull()
		}

		state.Organizations = append(state.Organizations, orgModel)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
