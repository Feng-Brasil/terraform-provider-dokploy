package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/feng-brasil/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ApiKeyResource{}
var _ resource.ResourceWithImportState = &ApiKeyResource{}

func NewApiKeyResource() resource.Resource {
	return &ApiKeyResource{}
}

type ApiKeyResource struct {
	client *client.DokployClient
}

type ApiKeyResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Key                 types.String `tfsdk:"key"`
	Start               types.String `tfsdk:"start"`
	UserID              types.String `tfsdk:"user_id"`
	OrganizationID      types.String `tfsdk:"organization_id"`
	ExpiresIn           types.Int64  `tfsdk:"expires_in"`
	ExpiresAt           types.String `tfsdk:"expires_at"`
	RateLimitEnabled    types.Bool   `tfsdk:"rate_limit_enabled"`
	RateLimitMax        types.Int64  `tfsdk:"rate_limit_max"`
	RateLimitTimeWindow types.Int64  `tfsdk:"rate_limit_time_window"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	CreatedAt           types.String `tfsdk:"created_at"`
}

func (r *ApiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *ApiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an API key in Dokploy. The API key value is only available after creation and stored in state as sensitive.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The actual API key value. Only available after creation and stored securely in state.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"start": schema.StringAttribute{
				Computed:    true,
				Description: "The first few characters of the API key for identification.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the user who owns this API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The organization ID to associate the API key with. If not specified, uses the current organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"expires_in": schema.Int64Attribute{
				Optional:    true,
				Description: "Time in seconds until the API key expires. Minimum is 86400 (1 day). If not set, the key does not expire.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"expires_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the API key expires.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"rate_limit_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether rate limiting is enabled for this API key. Defaults to true.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"rate_limit_max": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(10),
				Description: "Maximum number of requests allowed within the time window. Defaults to 10.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"rate_limit_time_window": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(86400000),
				Description: "Time window in milliseconds for rate limiting. Defaults to 86400000 (24 hours).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the API key is enabled.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the API key was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ApiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ApiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApiKeyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get organization ID - if not provided, get from current user
	var orgID string
	if !plan.OrganizationID.IsNull() && !plan.OrganizationID.IsUnknown() {
		orgID = plan.OrganizationID.ValueString()
	} else {
		user, err := r.client.GetUser()
		if err != nil {
			resp.Diagnostics.AddError("Error getting current user", err.Error())
			return
		}
		orgID = user.OrganizationID
	}

	input := client.ApiKeyCreateInput{
		Name: plan.Name.ValueString(),
		Metadata: map[string]string{
			"organizationId": orgID,
		},
	}

	if !plan.ExpiresIn.IsNull() && !plan.ExpiresIn.IsUnknown() {
		expiresIn := plan.ExpiresIn.ValueInt64()
		input.ExpiresIn = &expiresIn
	}

	if !plan.RateLimitEnabled.IsNull() && !plan.RateLimitEnabled.IsUnknown() {
		rateLimitEnabled := plan.RateLimitEnabled.ValueBool()
		input.RateLimitEnabled = &rateLimitEnabled
	}

	if !plan.RateLimitMax.IsNull() && !plan.RateLimitMax.IsUnknown() {
		rateLimitMax := int(plan.RateLimitMax.ValueInt64())
		input.RateLimitMax = &rateLimitMax
	}

	if !plan.RateLimitTimeWindow.IsNull() && !plan.RateLimitTimeWindow.IsUnknown() {
		rateLimitTimeWindow := plan.RateLimitTimeWindow.ValueInt64()
		input.RateLimitTimeWindow = &rateLimitTimeWindow
	}

	apiKey, err := r.client.CreateApiKey(input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating API key", err.Error())
		return
	}

	plan.ID = types.StringValue(apiKey.ID)
	plan.Name = types.StringValue(apiKey.Name)
	plan.Key = types.StringValue(apiKey.Key)
	plan.Start = types.StringValue(apiKey.Start)
	plan.UserID = types.StringValue(apiKey.UserID)
	plan.OrganizationID = types.StringValue(orgID)
	plan.Enabled = types.BoolValue(apiKey.Enabled)
	plan.RateLimitEnabled = types.BoolValue(apiKey.RateLimitEnabled)
	plan.RateLimitMax = types.Int64Value(int64(apiKey.RateLimitMax))
	plan.RateLimitTimeWindow = types.Int64Value(apiKey.RateLimitTimeWindow)
	plan.CreatedAt = types.StringValue(apiKey.CreatedAt)

	if apiKey.ExpiresAt != nil {
		plan.ExpiresAt = types.StringValue(*apiKey.ExpiresAt)
	} else {
		plan.ExpiresAt = types.StringNull()
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ApiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApiKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey, err := r.client.GetApiKeyByID(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading API key", err.Error())
		return
	}

	// Update state with current values from API
	// Note: The actual key value is NOT returned on read, so we preserve it from state
	state.Name = types.StringValue(apiKey.Name)
	state.Start = types.StringValue(apiKey.Start)
	state.UserID = types.StringValue(apiKey.UserID)
	state.Enabled = types.BoolValue(apiKey.Enabled)
	state.RateLimitEnabled = types.BoolValue(apiKey.RateLimitEnabled)
	state.RateLimitMax = types.Int64Value(int64(apiKey.RateLimitMax))
	state.RateLimitTimeWindow = types.Int64Value(apiKey.RateLimitTimeWindow)
	state.CreatedAt = types.StringValue(apiKey.CreatedAt)

	if apiKey.ExpiresAt != nil {
		state.ExpiresAt = types.StringValue(*apiKey.ExpiresAt)
	} else {
		state.ExpiresAt = types.StringNull()
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *ApiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// API keys are immutable - all changes require replacement
	// This is handled by RequiresReplace plan modifiers
	resp.Diagnostics.AddError(
		"API Key Update Not Supported",
		"API keys cannot be updated in place. Any changes require creating a new API key.",
	)
}

func (r *ApiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApiKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteApiKey(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError("Error deleting API key", err.Error())
		return
	}
}

func (r *ApiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Warn user that the key value cannot be imported
	resp.Diagnostics.AddWarning(
		"API Key Value Not Available",
		"The actual API key value is only available when the key is first created. "+
			"After import, the 'key' attribute will be empty. If you need the key value, "+
			"you must create a new API key.",
	)
}
