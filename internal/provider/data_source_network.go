package provider

import (
	"context"
	"fmt"

	"github.com/feng-brasil/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &NetworkDataSource{}

func NewNetworkDataSource() datasource.DataSource {
	return &NetworkDataSource{}
}

type NetworkDataSource struct {
	client *client.DokployClient
}

type NetworkDataSourceModel struct {
	ID             types.String      `tfsdk:"id"`
	Name           types.String      `tfsdk:"name"`
	Driver         types.String      `tfsdk:"driver"`
	Internal       types.Bool        `tfsdk:"internal"`
	Attachable     types.Bool        `tfsdk:"attachable"`
	EnableIPv4     types.Bool        `tfsdk:"enable_ipv4"`
	EnableIPv6     types.Bool        `tfsdk:"enable_ipv6"`
	MTU            types.Int64       `tfsdk:"mtu"`
	IPAM           *NetworkIPAMModel `tfsdk:"ipam"`
	ServerID       types.String      `tfsdk:"server_id"`
	OrganizationID types.String      `tfsdk:"organization_id"`
	CreatedAt      types.String      `tfsdk:"created_at"`
}

func (d *NetworkDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (d *NetworkDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single Dokploy network by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The network ID.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Network name.",
			},
			"driver": schema.StringAttribute{
				Computed:    true,
				Description: "Docker network driver.",
			},
			"internal": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the network is internal.",
			},
			"attachable": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether standalone containers can attach to the network.",
			},
			"enable_ipv4": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether IPv4 is enabled.",
			},
			"enable_ipv6": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether IPv6 is enabled.",
			},
			"mtu": schema.Int64Attribute{
				Computed:    true,
				Description: "Network MTU if configured.",
			},
			"ipam": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "IPAM settings for the network.",
				Attributes: map[string]schema.Attribute{
					"driver": schema.StringAttribute{
						Computed:    true,
						Description: "IPAM driver.",
					},
					"config": schema.ListNestedAttribute{
						Computed:    true,
						Description: "IPAM config entries.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"subnet": schema.StringAttribute{
									Computed:    true,
									Description: "Subnet CIDR.",
								},
								"gateway": schema.StringAttribute{
									Computed:    true,
									Description: "Gateway IP.",
								},
								"ip_range": schema.StringAttribute{
									Computed:    true,
									Description: "IP range.",
								},
							},
						},
					},
				},
			},
			"server_id": schema.StringAttribute{
				Computed:    true,
				Description: "Associated Dokploy server ID, when remote.",
			},
			"organization_id": schema.StringAttribute{
				Computed:    true,
				Description: "Organization ID that owns this network.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp.",
			},
		},
	}
}

func (d *NetworkDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.DokployClient, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *NetworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NetworkDataSourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	network, err := d.client.GetNetwork(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read network", err.Error())
		return
	}

	setNetworkDataSourceModelFromAPI(&data, network)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func setNetworkDataSourceModelFromAPI(model *NetworkDataSourceModel, network *client.Network) {
	model.ID = types.StringValue(network.NetworkID)
	model.Name = types.StringValue(network.Name)
	model.Driver = types.StringValue(network.Driver)
	model.Internal = types.BoolValue(network.Internal)
	model.Attachable = types.BoolValue(network.Attachable)
	model.EnableIPv4 = types.BoolValue(network.EnableIPv4)
	model.EnableIPv6 = types.BoolValue(network.EnableIPv6)
	model.OrganizationID = nullableString(network.OrganizationID)
	model.CreatedAt = nullableString(network.CreatedAt)
	model.IPAM = flattenNetworkIPAMModel(network.IPAM)

	if network.MTU != nil {
		model.MTU = types.Int64Value(*network.MTU)
	} else {
		model.MTU = types.Int64Null()
	}

	if network.ServerID != nil {
		model.ServerID = types.StringValue(*network.ServerID)
	} else {
		model.ServerID = types.StringNull()
	}
}
