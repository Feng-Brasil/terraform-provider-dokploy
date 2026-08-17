package provider

import (
	"context"
	"fmt"

	"github.com/feng-brasil/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &BackupFilesDataSource{}

func NewBackupFilesDataSource() datasource.DataSource {
	return &BackupFilesDataSource{}
}

type BackupFilesDataSource struct {
	client *client.DokployClient
}

type BackupFilesDataSourceModel struct {
	DestinationID types.String      `tfsdk:"destination_id"`
	Search        types.String      `tfsdk:"search"`
	ServerID      types.String      `tfsdk:"server_id"`
	Files         []BackupFileModel `tfsdk:"files"`
}

type BackupFileModel struct {
	Key          types.String `tfsdk:"key"`
	LastModified types.String `tfsdk:"last_modified"`
	Size         types.Int64  `tfsdk:"size"`
	ETag         types.String `tfsdk:"etag"`
	StorageClass types.String `tfsdk:"storage_class"`
}

func (d *BackupFilesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup_files"
}

func (d *BackupFilesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of backup files from a destination storage.",
		Attributes: map[string]schema.Attribute{
			"destination_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the backup destination to list files from.",
			},
			"search": schema.StringAttribute{
				Required:    true,
				Description: "Search prefix to filter backup files (e.g., backup prefix).",
			},
			"server_id": schema.StringAttribute{
				Optional:    true,
				Description: "Optional server ID to filter backups by server.",
			},
			"files": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of backup files.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Computed:    true,
							Description: "The key (path) of the backup file in storage.",
						},
						"last_modified": schema.StringAttribute{
							Computed:    true,
							Description: "The last modification timestamp of the file.",
						},
						"size": schema.Int64Attribute{
							Computed:    true,
							Description: "The size of the file in bytes.",
						},
						"etag": schema.StringAttribute{
							Computed:    true,
							Description: "The ETag of the file.",
						},
						"storage_class": schema.StringAttribute{
							Computed:    true,
							Description: "The storage class of the file.",
						},
					},
				},
			},
		},
	}
}

func (d *BackupFilesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BackupFilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config BackupFilesDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	destinationID := config.DestinationID.ValueString()
	search := config.Search.ValueString()
	serverID := ""
	if !config.ServerID.IsNull() && !config.ServerID.IsUnknown() {
		serverID = config.ServerID.ValueString()
	}

	files, err := d.client.ListBackupFiles(destinationID, search, serverID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to List Backup Files", err.Error())
		return
	}

	var state BackupFilesDataSourceModel
	state.DestinationID = config.DestinationID
	state.Search = config.Search
	state.ServerID = config.ServerID

	for _, file := range files {
		fileModel := BackupFileModel{
			Key:          types.StringValue(file.Key),
			LastModified: types.StringValue(file.LastModified),
			Size:         types.Int64Value(file.Size),
			ETag:         types.StringValue(file.ETag),
			StorageClass: types.StringValue(file.StorageClass),
		}
		state.Files = append(state.Files, fileModel)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
