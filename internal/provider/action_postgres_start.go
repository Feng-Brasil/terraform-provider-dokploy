package provider

import (
	"context"
	"fmt"

	"github.com/feng-brasil/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ action.Action = &PostgresStartAction{}
var _ action.ActionWithConfigure = &PostgresStartAction{}

func NewPostgresStartAction() action.Action {
	return &PostgresStartAction{}
}

type PostgresStartAction struct {
	client *client.DokployClient
}

type PostgresStartActionModel struct {
	PostgresID types.String `tfsdk:"postgres_id"`
}

func (a *PostgresStartAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgres_start"
}

func (a *PostgresStartAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Deploys and starts a Dokploy PostgreSQL service when it is stopped. A running service is left as it is. postgres.start scales the service to one replica, so it runs only when the configured replica count is 1.",
		Attributes: map[string]schema.Attribute{
			"postgres_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the dokploy_postgres instance to deploy and start.",
			},
		},
	}
}

func (a *PostgresStartAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Action Configure Type", fmt.Sprintf("Expected *client.DokployClient, got: %T", req.ProviderData))
		return
	}
	a.client = c
}

func (a *PostgresStartAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config PostgresStartActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	postgresID := config.PostgresID.ValueString()
	postgres, err := a.client.GetPostgres(postgresID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading PostgreSQL instance", err.Error())
		return
	}

	switch postgres.ApplicationStatus {
	case "done", "running":
		sendActionProgress(resp, fmt.Sprintf("PostgreSQL %s is already %s", postgresID, postgres.ApplicationStatus))
		return
	}

	sendActionProgress(resp, fmt.Sprintf("Deploying PostgreSQL %s", postgresID))
	if err := a.client.DeployPostgres(postgresID); err != nil {
		resp.Diagnostics.AddError("Error deploying PostgreSQL instance", err.Error())
		return
	}

	replicas := postgres.Replicas
	if replicas <= 0 {
		replicas = 1
	}
	if replicas > 1 {
		sendActionProgress(resp, fmt.Sprintf("PostgreSQL %s deployed with %d replicas", postgresID, replicas))
		return
	}

	sendActionProgress(resp, fmt.Sprintf("Starting PostgreSQL %s", postgresID))
	if err := a.client.StartPostgres(postgresID); err != nil {
		resp.Diagnostics.AddError("Error starting PostgreSQL instance", err.Error())
		return
	}
}

func sendActionProgress(resp *action.InvokeResponse, message string) {
	if resp.SendProgress == nil {
		return
	}
	resp.SendProgress(action.InvokeProgressEvent{Message: message})
}
