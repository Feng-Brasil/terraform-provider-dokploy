package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/feng-brasil/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var destinationFlagPattern = regexp.MustCompile(`^--[a-zA-Z0-9-]+(=[a-zA-Z0-9._:/@-]+)?$`)

var _ resource.Resource = &DestinationResource{}
var _ resource.ResourceWithImportState = &DestinationResource{}

func NewDestinationResource() resource.Resource {
	return &DestinationResource{}
}

type DestinationResource struct {
	client *client.DokployClient
}

type DestinationResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	StorageProvider types.String `tfsdk:"storage_provider"`
	AccessKey       types.String `tfsdk:"access_key"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	Bucket          types.String `tfsdk:"bucket"`
	Region          types.String `tfsdk:"region"`
	Endpoint        types.String `tfsdk:"endpoint"`
	AdditionalFlags types.List   `tfsdk:"additional_flags"`
	ServerID        types.String `tfsdk:"server_id"`
}

func (r *DestinationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_destination"
}

func (r *DestinationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a backup destination (S3, MinIO, etc.) for Dokploy backups.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the destination",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the destination",
			},
			"storage_provider": schema.StringAttribute{
				Required:    true,
				Description: "Storage provider type (e.g., 's3', 'minio')",
			},
			"access_key": schema.StringAttribute{
				Required:    true,
				Description: "Access key for the storage provider",
			},
			"secret_access_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Secret access key for the storage provider",
			},
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "Bucket name for storing backups",
			},
			"region": schema.StringAttribute{
				Required:    true,
				Description: "Region where the bucket is located",
			},
			"endpoint": schema.StringAttribute{
				Required:    true,
				Description: "Endpoint URL for the storage provider",
			},
			"additional_flags": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Extra rclone flags sent to destination.create and destination.update. The API requires this field, so an omitted value is sent as an empty array. Each flag must start with --, for example --s3-sign-accept-encoding=false.",
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.RegexMatches(
						destinationFlagPattern,
						"must start with -- and contain only letters, numbers, and . _ : / @ - (for example --s3-sign-accept-encoding=false)",
					)),
				},
			},
			"server_id": schema.StringAttribute{
				Optional:    true,
				Description: "Optional server ID accepted by the destination API for remote execution. The destination record itself does not persist this column; the configured value is kept in state.",
			},
		},
	}
}

func (r *DestinationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DestinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DestinationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dest, diags := destinationFromPlan(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createdDest, err := r.client.CreateDestination(dest)
	if err != nil {
		resp.Diagnostics.AddError("Error creating destination", err.Error())
		return
	}

	// Keep configured values. The API echoes the same fields, and replacing a
	// null additional_flags with [] would fail the apply consistency check.
	plan.ID = types.StringValue(createdDest.DestinationID)
	plan.ServerID = preserveServerID(plan.ServerID, createdDest.ServerID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DestinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DestinationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dest, err := r.client.GetDestination(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading destination", err.Error())
		return
	}

	state.Name = types.StringValue(dest.Name)
	state.StorageProvider = types.StringValue(dest.Provider)
	state.AccessKey = types.StringValue(dest.AccessKey)
	state.Bucket = types.StringValue(dest.Bucket)
	state.Region = types.StringValue(dest.Region)
	state.Endpoint = types.StringValue(dest.Endpoint)
	state.AdditionalFlags = additionalFlagsState(ctx, state.AdditionalFlags, dest.AdditionalFlags)
	state.ServerID = preserveServerID(state.ServerID, dest.ServerID)
	// Don't update secret_access_key from API response

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *DestinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DestinationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dest, diags := destinationFromPlan(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	dest.DestinationID = plan.ID.ValueString()

	updatedDest, err := r.client.UpdateDestination(dest)
	if err != nil {
		resp.Diagnostics.AddError("Error updating destination", err.Error())
		return
	}

	plan.Name = types.StringValue(updatedDest.Name)
	plan.StorageProvider = types.StringValue(updatedDest.Provider)
	plan.AccessKey = types.StringValue(updatedDest.AccessKey)
	plan.Bucket = types.StringValue(updatedDest.Bucket)
	plan.Region = types.StringValue(updatedDest.Region)
	plan.Endpoint = types.StringValue(updatedDest.Endpoint)
	plan.AdditionalFlags = additionalFlagsState(ctx, plan.AdditionalFlags, updatedDest.AdditionalFlags)
	plan.ServerID = preserveServerID(plan.ServerID, updatedDest.ServerID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DestinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DestinationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDestination(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") || strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError("Error deleting destination", err.Error())
		return
	}
}

func (r *DestinationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func destinationFromPlan(ctx context.Context, plan *DestinationResourceModel) (client.Destination, diag.Diagnostics) {
	dest := client.Destination{
		Name:            plan.Name.ValueString(),
		Provider:        plan.StorageProvider.ValueString(),
		AccessKey:       plan.AccessKey.ValueString(),
		SecretAccessKey: plan.SecretAccessKey.ValueString(),
		Bucket:          plan.Bucket.ValueString(),
		Region:          plan.Region.ValueString(),
		Endpoint:        plan.Endpoint.ValueString(),
		AdditionalFlags: []string{},
	}
	if flags, ok, diags := stringList(ctx, plan.AdditionalFlags); ok {
		dest.AdditionalFlags = flags
	} else if diags.HasError() {
		return dest, diags
	}
	if !plan.ServerID.IsNull() && !plan.ServerID.IsUnknown() && plan.ServerID.ValueString() != "" {
		serverID := plan.ServerID.ValueString()
		dest.ServerID = &serverID
	}
	return dest, nil
}

func preserveServerID(current types.String, api *string) types.String {
	if api != nil && *api != "" {
		return types.StringValue(*api)
	}
	if !current.IsNull() && !current.IsUnknown() {
		return current
	}
	return types.StringNull()
}

func additionalFlagsState(ctx context.Context, current types.List, flags []string) types.List {
	if len(flags) == 0 {
		if current.IsNull() || current.IsUnknown() {
			return types.ListNull(types.StringType)
		}
		return current
	}
	return stringListOrNull(ctx, flags)
}
