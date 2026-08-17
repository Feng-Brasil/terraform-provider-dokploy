package provider

import (
	"context"
	"fmt"

	"github.com/feng-brasil/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &VolumeBackupsDataSource{}

func NewVolumeBackupsDataSource() datasource.DataSource {
	return &VolumeBackupsDataSource{}
}

type VolumeBackupsDataSource struct {
	client *client.DokployClient
}

type VolumeBackupsDataSourceModel struct {
	ServiceID     types.String        `tfsdk:"service_id"`
	ServiceType   types.String        `tfsdk:"service_type"`
	VolumeBackups []VolumeBackupModel `tfsdk:"volume_backups"`
}

type VolumeBackupModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	VolumeName      types.String `tfsdk:"volume_name"`
	Prefix          types.String `tfsdk:"prefix"`
	DestinationID   types.String `tfsdk:"destination_id"`
	CronExpression  types.String `tfsdk:"cron_expression"`
	ServiceType     types.String `tfsdk:"service_type"`
	AppName         types.String `tfsdk:"app_name"`
	ServiceName     types.String `tfsdk:"service_name"`
	TurnOff         types.Bool   `tfsdk:"turn_off"`
	KeepLatestCount types.Int64  `tfsdk:"keep_latest_count"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	CreatedAt       types.String `tfsdk:"created_at"`
}

func (d *VolumeBackupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume_backups"
}

func (d *VolumeBackupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of volume backups for a specific service in Dokploy.",
		Attributes: map[string]schema.Attribute{
			"service_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the service to list volume backups for.",
			},
			"service_type": schema.StringAttribute{
				Required:    true,
				Description: "Type of service: application, postgres, mysql, mariadb, mongo, redis, or compose.",
				Validators: []validator.String{
					stringvalidator.OneOf("application", "postgres", "mysql", "mariadb", "mongo", "redis", "compose"),
				},
			},
			"volume_backups": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of volume backups for the service.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique identifier of the volume backup.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the volume backup configuration.",
						},
						"volume_name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the Docker volume being backed up.",
						},
						"prefix": schema.StringAttribute{
							Computed:    true,
							Description: "Prefix for backup files in the destination storage.",
						},
						"destination_id": schema.StringAttribute{
							Computed:    true,
							Description: "ID of the backup destination.",
						},
						"cron_expression": schema.StringAttribute{
							Computed:    true,
							Description: "Cron schedule for backups.",
						},
						"service_type": schema.StringAttribute{
							Computed:    true,
							Description: "Type of service being backed up.",
						},
						"app_name": schema.StringAttribute{
							Computed:    true,
							Description: "Docker app/service name.",
						},
						"service_name": schema.StringAttribute{
							Computed:    true,
							Description: "Service name within a compose stack.",
						},
						"turn_off": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the service is stopped during backup.",
						},
						"keep_latest_count": schema.Int64Attribute{
							Computed:    true,
							Description: "Number of recent backups to keep.",
						},
						"enabled": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the backup schedule is enabled.",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "Creation timestamp of the volume backup.",
						},
					},
				},
			},
		},
	}
}

func (d *VolumeBackupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VolumeBackupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config VolumeBackupsDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := config.ServiceID.ValueString()
	serviceType := config.ServiceType.ValueString()

	backups, err := d.client.ListVolumeBackups(serviceID, serviceType)
	if err != nil {
		resp.Diagnostics.AddError("Unable to List Volume Backups", err.Error())
		return
	}

	var state VolumeBackupsDataSourceModel
	state.ServiceID = config.ServiceID
	state.ServiceType = config.ServiceType

	for _, backup := range backups {
		backupModel := VolumeBackupModel{
			ID:              types.StringValue(backup.VolumeBackupID),
			Name:            types.StringValue(backup.Name),
			VolumeName:      types.StringValue(backup.VolumeName),
			Prefix:          types.StringValue(backup.Prefix),
			DestinationID:   types.StringValue(backup.DestinationID),
			CronExpression:  types.StringValue(backup.CronExpression),
			ServiceType:     types.StringValue(backup.ServiceType),
			AppName:         types.StringValue(backup.AppName),
			TurnOff:         types.BoolValue(backup.TurnOff),
			KeepLatestCount: types.Int64Value(int64(backup.KeepLatestCount)),
			Enabled:         types.BoolValue(backup.Enabled),
			CreatedAt:       types.StringValue(backup.CreatedAt),
		}

		if backup.ServiceName != nil && *backup.ServiceName != "" {
			backupModel.ServiceName = types.StringValue(*backup.ServiceName)
		} else {
			backupModel.ServiceName = types.StringNull()
		}

		state.VolumeBackups = append(state.VolumeBackups, backupModel)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
