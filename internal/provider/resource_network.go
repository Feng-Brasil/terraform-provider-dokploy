package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/feng-brasil/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &NetworkResource{}
var _ resource.ResourceWithImportState = &NetworkResource{}

func NewNetworkResource() resource.Resource {
	return &NetworkResource{}
}

type NetworkResource struct {
	client *client.DokployClient
}

type NetworkResourceModel struct {
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

func (r *NetworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (r *NetworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Docker networks in Dokploy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the network.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the Docker network.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"driver": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("bridge"),
				Description: "Docker network driver.",
				Validators: []validator.String{
					stringvalidator.OneOf("bridge", "overlay"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"internal": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the network is internal.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"attachable": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether standalone containers can attach to this network.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"enable_ipv4": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether IPv4 is enabled.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"enable_ipv6": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether IPv6 is enabled.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"mtu": schema.Int64Attribute{
				Optional:    true,
				Description: "Optional MTU for the network.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"ipam": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "IPAM settings for the network.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"driver": schema.StringAttribute{
						Optional:    true,
						Description: "IPAM driver.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"config": schema.ListNestedAttribute{
						Optional:    true,
						Description: "IPAM config entries.",
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplace(),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"subnet": schema.StringAttribute{
									Optional:    true,
									Description: "Subnet CIDR for the network.",
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.RequiresReplace(),
									},
								},
								"gateway": schema.StringAttribute{
									Optional:    true,
									Description: "Gateway IP for the subnet.",
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.RequiresReplace(),
									},
								},
								"ip_range": schema.StringAttribute{
									Optional:    true,
									Description: "IP range within the subnet.",
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.RequiresReplace(),
									},
								},
							},
						},
					},
				},
			},
			"server_id": schema.StringAttribute{
				Optional:    true,
				Description: "Optional target Dokploy server ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

func (r *NetworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.DokployClient, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *NetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NetworkResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	network := client.Network{
		Name:       plan.Name.ValueString(),
		Driver:     plan.Driver.ValueString(),
		Internal:   plan.Internal.ValueBool(),
		Attachable: plan.Attachable.ValueBool(),
		EnableIPv4: plan.EnableIPv4.ValueBool(),
		EnableIPv6: plan.EnableIPv6.ValueBool(),
		IPAM:       expandNetworkIPAMModel(plan.IPAM),
	}

	if !plan.MTU.IsNull() && !plan.MTU.IsUnknown() {
		mtu := plan.MTU.ValueInt64()
		network.MTU = &mtu
	}
	if !plan.ServerID.IsNull() && !plan.ServerID.IsUnknown() {
		serverID := plan.ServerID.ValueString()
		network.ServerID = &serverID
	}

	createdNetwork, err := r.client.CreateNetwork(network)
	if err != nil {
		resp.Diagnostics.AddError("Error creating network", err.Error())
		return
	}

	setNetworkResourceModelFromAPI(&plan, createdNetwork)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *NetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NetworkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	network, err := r.client.GetNetwork(state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading network", err.Error())
		return
	}

	setNetworkResourceModelFromAPI(&state, network)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state NetworkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	network, err := r.client.GetNetwork(state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading network during update", err.Error())
		return
	}

	setNetworkResourceModelFromAPI(&state, network)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *NetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NetworkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNetwork(state.ID.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting network", err.Error())
	}
}

func (r *NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func setNetworkResourceModelFromAPI(model *NetworkResourceModel, network *client.Network) {
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
