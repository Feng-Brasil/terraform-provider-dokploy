package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/feng-brasil/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &PostgresResource{}
var _ resource.ResourceWithImportState = &PostgresResource{}

func NewPostgresResource() resource.Resource {
	return &PostgresResource{}
}

type PostgresResource struct {
	client *client.DokployClient
}

type PostgresResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	AppName           types.String `tfsdk:"app_name"`
	ResolvedAppName   types.String `tfsdk:"resolved_app_name"`
	Description       types.String `tfsdk:"description"`
	DatabaseName      types.String `tfsdk:"database_name"`
	DatabaseUser      types.String `tfsdk:"database_user"`
	DatabasePassword  types.String `tfsdk:"database_password"`
	DockerImage       types.String `tfsdk:"docker_image"`
	Command           types.String `tfsdk:"command"`
	Args              types.List   `tfsdk:"args"`
	Env               types.String `tfsdk:"env"`
	MemoryReservation types.String `tfsdk:"memory_reservation"`
	MemoryLimit       types.String `tfsdk:"memory_limit"`
	CPUReservation    types.String `tfsdk:"cpu_reservation"`
	CPULimit          types.String `tfsdk:"cpu_limit"`
	ExternalPort      types.Int64  `tfsdk:"external_port"`
	EnvironmentID     types.String `tfsdk:"environment_id"`
	ApplicationStatus types.String `tfsdk:"application_status"`
	Replicas          types.Int64  `tfsdk:"replicas"`
	ServerID          types.String `tfsdk:"server_id"`
	DeployOnCreate    types.Bool   `tfsdk:"deploy_on_create"`
	DeployOnUpdate    types.Bool   `tfsdk:"deploy_on_update"`

	HealthCheckSwarm     types.String `tfsdk:"health_check_swarm"`
	RestartPolicySwarm   types.String `tfsdk:"restart_policy_swarm"`
	PlacementSwarm       types.String `tfsdk:"placement_swarm"`
	UpdateConfigSwarm    types.String `tfsdk:"update_config_swarm"`
	RollbackConfigSwarm  types.String `tfsdk:"rollback_config_swarm"`
	ModeSwarm            types.String `tfsdk:"mode_swarm"`
	LabelsSwarm          types.String `tfsdk:"labels_swarm"`
	NetworkSwarm         types.String `tfsdk:"network_swarm"`
	StopGracePeriodSwarm types.Int64  `tfsdk:"stop_grace_period_swarm"`
	EndpointSpecSwarm    types.String `tfsdk:"endpoint_spec_swarm"`
	UlimitsSwarm         types.String `tfsdk:"ulimits_swarm"`
	NetworkIDs           types.List   `tfsdk:"network_ids"`
	DetachDokployNetwork types.Bool   `tfsdk:"detach_dokploy_network"`
}

func (r *PostgresResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgres"
}

func (r *PostgresResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PostgreSQL database instance in Dokploy. Create calls postgres.create, then postgres.update with the full configuration. postgres.deploy and postgres.start run on create only when deploy_on_create is true. Updates call postgres.update and, by default, postgres.deploy so Swarm settings take effect.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the PostgreSQL instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the PostgreSQL instance.",
			},
			"app_name": schema.StringAttribute{
				Required:    true,
				Description: "Application name prefix. Dokploy appends a random suffix and keeps that generated name as the Swarm service name. The configured prefix is stored in state.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resolved_app_name": schema.StringAttribute{
				Computed:    true,
				Description: "Swarm service name Dokploy generated from app_name, including the random suffix.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the PostgreSQL instance.",
			},
			"database_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the database to create.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database_user": schema.StringAttribute{
				Required:    true,
				Description: "Database user name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database_password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Password for the database user. On update, the provider also calls postgres.changePassword when a container is already running.",
			},
			"docker_image": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Docker image to use. Dokploy defaults to postgres:18 when this is omitted.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"command": schema.StringAttribute{
				Optional:    true,
				Description: "Custom command to run in the container.",
			},
			"args": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Arguments passed to the container command.",
			},
			"env": schema.StringAttribute{
				Optional:    true,
				Description: "Environment variables for the container, in KEY=VALUE lines.",
			},
			"memory_reservation": schema.StringAttribute{
				Optional:    true,
				Description: "Memory reservation for the container.",
			},
			"memory_limit": schema.StringAttribute{
				Optional:    true,
				Description: "Memory limit for the container.",
			},
			"cpu_reservation": schema.StringAttribute{
				Optional:    true,
				Description: "CPU reservation for the container.",
			},
			"cpu_limit": schema.StringAttribute{
				Optional:    true,
				Description: "CPU limit for the container.",
			},
			"external_port": schema.Int64Attribute{
				Optional:    true,
				Description: "External port to expose the PostgreSQL instance.",
			},
			"environment_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the environment to deploy the PostgreSQL instance in.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"application_status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of the PostgreSQL application (idle, running, done, error).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"replicas": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of replicas for the PostgreSQL instance.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"server_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the server to deploy the PostgreSQL instance on.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"deploy_on_create": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "When true, create calls postgres.deploy and then postgres.start after the instance is configured. Defaults to false so volumes and bind mounts can be attached before the service starts. Use the dokploy_postgres_start action to start it later.",
			},
			"deploy_on_update": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "When true (default), an update calls postgres.deploy after postgres.update so Swarm, resource, and image changes are applied. Set to false to store the configuration without redeploying.",
			},
			"health_check_swarm": schema.StringAttribute{
				Optional:    true,
				Description: "Docker Swarm health check (JSON object). Keys match the Dokploy schema: Test, Interval, Timeout, StartPeriod, Retries.",
			},
			"restart_policy_swarm": schema.StringAttribute{
				Optional:    true,
				Description: "Docker Swarm restart policy (JSON object). Keys: Condition, Delay, MaxAttempts, Window.",
			},
			"placement_swarm": schema.StringAttribute{
				Optional:    true,
				Description: "Docker Swarm placement (JSON object). Keys: Constraints, Preferences, MaxReplicas, Platforms.",
			},
			"update_config_swarm": schema.StringAttribute{
				Optional:    true,
				Description: "Docker Swarm update config (JSON object). Keys: Parallelism, Delay, FailureAction, Monitor, MaxFailureRatio, Order.",
			},
			"rollback_config_swarm": schema.StringAttribute{
				Optional:    true,
				Description: "Docker Swarm rollback config (JSON object). Same keys as update_config_swarm.",
			},
			"mode_swarm": schema.StringAttribute{
				Optional:    true,
				Description: "Docker Swarm service mode (JSON object). Keys: Replicated, Global, ReplicatedJob, GlobalJob.",
			},
			"labels_swarm": schema.StringAttribute{
				Optional:    true,
				Description: "Docker Swarm service labels (JSON object of string keys and values).",
			},
			"network_swarm": schema.StringAttribute{
				Optional:    true,
				Description: "Docker Swarm networks (JSON array). Each item may set Target, Aliases, and DriverOpts.",
			},
			"stop_grace_period_swarm": schema.Int64Attribute{
				Optional:    true,
				Description: "Stop grace period in nanoseconds for Docker Swarm mode.",
			},
			"endpoint_spec_swarm": schema.StringAttribute{
				Optional:    true,
				Description: "Docker Swarm endpoint spec (JSON object). Keys: Mode, Ports.",
			},
			"ulimits_swarm": schema.StringAttribute{
				Optional:    true,
				Description: "Docker Swarm ulimits (JSON array). Each item has Name, Soft, and Hard.",
			},
			"network_ids": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Dokploy network IDs to attach.",
			},
			"detach_dokploy_network": schema.BoolAttribute{
				Optional:    true,
				Description: "Detach the service from the default Dokploy network.",
			},
		},
	}
}

func (r *PostgresResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.DokployClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.DokployClient, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *PostgresResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PostgresResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreatePostgres(client.Postgres{
		Name:             plan.Name.ValueString(),
		AppName:          plan.AppName.ValueString(),
		Description:      plan.Description.ValueString(),
		DatabaseName:     plan.DatabaseName.ValueString(),
		DatabaseUser:     plan.DatabaseUser.ValueString(),
		DatabasePassword: plan.DatabasePassword.ValueString(),
		DockerImage:      plan.DockerImage.ValueString(),
		EnvironmentID:    plan.EnvironmentID.ValueString(),
		ServerID:         plan.ServerID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating PostgreSQL instance", err.Error())
		return
	}

	updateBody, diags := postgresUpdateFromPlan(ctx, &plan, created.PostgresID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.UpdatePostgres(updateBody); err != nil {
		resp.Diagnostics.AddError("Error updating PostgreSQL instance after creation", err.Error())
		return
	}

	if deployOnCreateEnabled(plan.DeployOnCreate) {
		// deploy creates the Swarm service. start only scales an existing
		// service to 1, so it runs after deploy and only when that scale
		// matches the configured replica count.
		if err := r.client.DeployPostgres(created.PostgresID); err != nil {
			resp.Diagnostics.AddWarning("PostgreSQL deploy failed", fmt.Sprintf("The instance was created and updated, but postgres.deploy failed: %s", err.Error()))
		} else if postgresReplicas(&plan) <= 1 {
			if err := r.client.StartPostgres(created.PostgresID); err != nil {
				resp.Diagnostics.AddWarning("PostgreSQL start failed", fmt.Sprintf("The instance was deployed, but postgres.start failed: %s", err.Error()))
			}
		}
	}

	finalPostgres, err := r.client.GetPostgres(created.PostgresID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading PostgreSQL instance after create", err.Error())
		return
	}
	fillComputedPostgres(&plan, finalPostgres)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PostgresResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PostgresResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	postgres, err := r.client.GetPostgres(state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading PostgreSQL instance", err.Error())
		return
	}

	readPostgresIntoState(ctx, &state, postgres)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *PostgresResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PostgresResourceModel
	var state PostgresResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateBody, diags := postgresUpdateFromPlan(ctx, &plan, plan.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.UpdatePostgres(updateBody); err != nil {
		resp.Diagnostics.AddError("Error updating PostgreSQL instance", err.Error())
		return
	}

	passwordChanged := plan.DatabasePassword.ValueString() != state.DatabasePassword.ValueString()
	if passwordChanged {
		if err := r.client.ChangePostgresPassword(plan.ID.ValueString(), plan.DatabasePassword.ValueString()); err != nil {
			resp.Diagnostics.AddWarning(
				"PostgreSQL password change failed",
				fmt.Sprintf("The stored password was updated, but postgres.changePassword failed. The live role is unchanged until a container is running: %s", err.Error()),
			)
		}
	}

	if plan.DeployOnUpdate.IsNull() || plan.DeployOnUpdate.IsUnknown() || plan.DeployOnUpdate.ValueBool() {
		if err := r.client.DeployPostgres(plan.ID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error deploying PostgreSQL instance", fmt.Sprintf("Configuration was updated, but postgres.deploy failed: %s", err.Error()))
			return
		}
	}

	updatedPostgres, err := r.client.GetPostgres(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading PostgreSQL instance after update", err.Error())
		return
	}
	fillComputedPostgres(&plan, updatedPostgres)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PostgresResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PostgresResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePostgres(state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Error deleting PostgreSQL instance", err.Error())
		return
	}
}

func (r *PostgresResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func deployOnCreateEnabled(value types.Bool) bool {
	return !value.IsNull() && !value.IsUnknown() && value.ValueBool()
}

func postgresReplicas(plan *PostgresResourceModel) int {
	if plan.Replicas.IsNull() || plan.Replicas.IsUnknown() {
		return 1
	}
	return int(plan.Replicas.ValueInt64())
}

func postgresUpdateFromPlan(ctx context.Context, plan *PostgresResourceModel, id string) (client.Postgres, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := client.Postgres{
		PostgresID: id,
		Name:       plan.Name.ValueString(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body.Description = plan.Description.ValueString()
		body.DescriptionSet = true
	}
	if !plan.DatabasePassword.IsNull() && !plan.DatabasePassword.IsUnknown() {
		body.DatabasePassword = plan.DatabasePassword.ValueString()
		body.DatabasePasswordSet = true
	}
	if !plan.DockerImage.IsNull() && !plan.DockerImage.IsUnknown() && plan.DockerImage.ValueString() != "" {
		body.DockerImage = plan.DockerImage.ValueString()
		body.DockerImageSet = true
	}
	if !plan.Command.IsNull() && !plan.Command.IsUnknown() {
		body.Command = plan.Command.ValueString()
		body.CommandSet = true
	}
	if !plan.Env.IsNull() && !plan.Env.IsUnknown() {
		body.Env = plan.Env.ValueString()
		body.EnvSet = true
	}
	if !plan.MemoryReservation.IsNull() && !plan.MemoryReservation.IsUnknown() {
		body.MemoryReservation = plan.MemoryReservation.ValueString()
		body.MemoryReservationSet = true
	}
	if !plan.MemoryLimit.IsNull() && !plan.MemoryLimit.IsUnknown() {
		body.MemoryLimit = plan.MemoryLimit.ValueString()
		body.MemoryLimitSet = true
	}
	if !plan.CPUReservation.IsNull() && !plan.CPUReservation.IsUnknown() {
		body.CPUReservation = plan.CPUReservation.ValueString()
		body.CPUReservationSet = true
	}
	if !plan.CPULimit.IsNull() && !plan.CPULimit.IsUnknown() {
		body.CPULimit = plan.CPULimit.ValueString()
		body.CPULimitSet = true
	}
	if !plan.ExternalPort.IsNull() && !plan.ExternalPort.IsUnknown() {
		body.ExternalPort = int(plan.ExternalPort.ValueInt64())
		body.ExternalPortSet = true
	}
	if !plan.Replicas.IsNull() && !plan.Replicas.IsUnknown() {
		body.Replicas = int(plan.Replicas.ValueInt64())
		body.ReplicasSet = true
	}
	if !plan.DetachDokployNetwork.IsNull() && !plan.DetachDokployNetwork.IsUnknown() {
		body.DetachDokployNetwork = plan.DetachDokployNetwork.ValueBool()
		body.DetachDokployNetworkSet = true
	}
	if !plan.StopGracePeriodSwarm.IsNull() && !plan.StopGracePeriodSwarm.IsUnknown() {
		grace := plan.StopGracePeriodSwarm.ValueInt64()
		body.StopGracePeriodSwarm = &grace
		body.StopGracePeriodSwarmSet = true
	}

	if args, ok, listDiags := stringList(ctx, plan.Args); ok {
		body.Args = args
		body.ArgsSet = true
	} else {
		diags.Append(listDiags...)
	}
	if ids, ok, listDiags := stringList(ctx, plan.NetworkIDs); ok {
		body.NetworkIds = ids
		body.NetworkIdsSet = true
	} else {
		diags.Append(listDiags...)
	}

	assignObject := func(attr types.String, field *map[string]interface{}, name string) {
		if attr.IsNull() || attr.IsUnknown() || attr.ValueString() == "" {
			return
		}
		parsed, err := parseJSONObject(attr.ValueString())
		if err != nil {
			diags.AddError("Invalid "+name, err.Error())
			return
		}
		*field = parsed
	}
	assignObject(plan.HealthCheckSwarm, &body.HealthCheckSwarm, "health_check_swarm")
	assignObject(plan.RestartPolicySwarm, &body.RestartPolicySwarm, "restart_policy_swarm")
	assignObject(plan.PlacementSwarm, &body.PlacementSwarm, "placement_swarm")
	assignObject(plan.UpdateConfigSwarm, &body.UpdateConfigSwarm, "update_config_swarm")
	assignObject(plan.RollbackConfigSwarm, &body.RollbackConfigSwarm, "rollback_config_swarm")
	assignObject(plan.ModeSwarm, &body.ModeSwarm, "mode_swarm")
	assignObject(plan.LabelsSwarm, &body.LabelsSwarm, "labels_swarm")
	assignObject(plan.EndpointSpecSwarm, &body.EndpointSpecSwarm, "endpoint_spec_swarm")

	if !plan.NetworkSwarm.IsNull() && !plan.NetworkSwarm.IsUnknown() && plan.NetworkSwarm.ValueString() != "" {
		parsed, err := parseJSONObjectList(plan.NetworkSwarm.ValueString())
		if err != nil {
			diags.AddError("Invalid network_swarm", err.Error())
		} else {
			body.NetworkSwarm = parsed
			body.NetworkSwarmSet = true
		}
	}
	if !plan.UlimitsSwarm.IsNull() && !plan.UlimitsSwarm.IsUnknown() && plan.UlimitsSwarm.ValueString() != "" {
		parsed, err := parseJSONObjectList(plan.UlimitsSwarm.ValueString())
		if err != nil {
			diags.AddError("Invalid ulimits_swarm", err.Error())
		} else {
			body.UlimitsSwarm = parsed
			body.UlimitsSwarmSet = true
		}
	}

	return body, diags
}

func fillComputedPostgres(state *PostgresResourceModel, postgres *client.Postgres) {
	state.ID = types.StringValue(postgres.PostgresID)
	if postgres.AppName != "" {
		state.ResolvedAppName = types.StringValue(postgres.AppName)
	} else if state.ResolvedAppName.IsUnknown() {
		state.ResolvedAppName = types.StringNull()
	}
	// Keep the configured prefix. Dokploy stores prefix + random suffix, and
	// writing that suffix back makes apply report an inconsistent result.
	if postgres.AppName != "" && (state.AppName.IsNull() || state.AppName.IsUnknown()) {
		state.AppName = types.StringValue(postgres.AppName)
	}
	if (state.DockerImage.IsNull() || state.DockerImage.IsUnknown()) && postgres.DockerImage != "" {
		state.DockerImage = types.StringValue(postgres.DockerImage)
	}
	if state.Replicas.IsNull() || state.Replicas.IsUnknown() {
		replicas := postgres.Replicas
		if replicas <= 0 {
			replicas = 1
		}
		state.Replicas = types.Int64Value(int64(replicas))
	}
	if postgres.ApplicationStatus != "" {
		state.ApplicationStatus = types.StringValue(postgres.ApplicationStatus)
	} else if state.ApplicationStatus.IsUnknown() {
		state.ApplicationStatus = types.StringNull()
	}
	if postgres.DatabasePassword != "" && (state.DatabasePassword.IsNull() || state.DatabasePassword.IsUnknown()) {
		state.DatabasePassword = types.StringValue(postgres.DatabasePassword)
	}
}

func readPostgresIntoState(ctx context.Context, state *PostgresResourceModel, postgres *client.Postgres) {
	configuredAppName := state.AppName
	deployOnCreate := state.DeployOnCreate
	deployOnUpdate := state.DeployOnUpdate
	password := state.DatabasePassword

	state.ID = types.StringValue(postgres.PostgresID)
	state.Name = types.StringValue(postgres.Name)
	state.EnvironmentID = types.StringValue(postgres.EnvironmentID)
	if postgres.DatabaseName != "" {
		state.DatabaseName = types.StringValue(postgres.DatabaseName)
	}
	if postgres.DatabaseUser != "" {
		state.DatabaseUser = types.StringValue(postgres.DatabaseUser)
	}
	if postgres.DatabasePassword != "" {
		state.DatabasePassword = types.StringValue(postgres.DatabasePassword)
	} else if !password.IsNull() && !password.IsUnknown() {
		state.DatabasePassword = password
	}

	if !configuredAppName.IsNull() && !configuredAppName.IsUnknown() {
		state.AppName = configuredAppName
	} else if postgres.AppName != "" {
		state.AppName = types.StringValue(postgres.AppName)
	}
	if postgres.AppName != "" {
		state.ResolvedAppName = types.StringValue(postgres.AppName)
	}
	if postgres.ApplicationStatus != "" {
		state.ApplicationStatus = types.StringValue(postgres.ApplicationStatus)
	}
	if postgres.DockerImage != "" {
		state.DockerImage = types.StringValue(postgres.DockerImage)
	}
	replicas := postgres.Replicas
	if replicas <= 0 {
		replicas = 1
	}
	state.Replicas = types.Int64Value(int64(replicas))

	state.Description = stringOrNull(postgres.Description)
	state.Command = stringOrNull(postgres.Command)
	state.Env = stringOrNull(postgres.Env)
	state.MemoryReservation = stringOrNull(postgres.MemoryReservation)
	state.MemoryLimit = stringOrNull(postgres.MemoryLimit)
	state.CPUReservation = stringOrNull(postgres.CPUReservation)
	state.CPULimit = stringOrNull(postgres.CPULimit)
	if postgres.ExternalPort > 0 {
		state.ExternalPort = types.Int64Value(int64(postgres.ExternalPort))
	} else {
		state.ExternalPort = types.Int64Null()
	}
	if postgres.ServerID != "" {
		state.ServerID = types.StringValue(postgres.ServerID)
	} else if state.ServerID.IsNull() || state.ServerID.IsUnknown() {
		state.ServerID = types.StringNull()
	}

	state.Args = stringListOrNull(ctx, postgres.Args)
	state.NetworkIDs = stringListOrNull(ctx, postgres.NetworkIds)
	if postgres.DetachDokployNetwork {
		state.DetachDokployNetwork = types.BoolValue(true)
	} else if state.DetachDokployNetwork.IsNull() || state.DetachDokployNetwork.IsUnknown() {
		state.DetachDokployNetwork = types.BoolNull()
	}
	if postgres.StopGracePeriodSwarm != nil {
		state.StopGracePeriodSwarm = types.Int64Value(*postgres.StopGracePeriodSwarm)
	} else {
		state.StopGracePeriodSwarm = types.Int64Null()
	}

	state.HealthCheckSwarm = jsonAttrFrom(state.HealthCheckSwarm, postgres.HealthCheckSwarm)
	state.RestartPolicySwarm = jsonAttrFrom(state.RestartPolicySwarm, postgres.RestartPolicySwarm)
	state.PlacementSwarm = jsonAttrFrom(state.PlacementSwarm, postgres.PlacementSwarm)
	state.UpdateConfigSwarm = jsonAttrFrom(state.UpdateConfigSwarm, postgres.UpdateConfigSwarm)
	state.RollbackConfigSwarm = jsonAttrFrom(state.RollbackConfigSwarm, postgres.RollbackConfigSwarm)
	state.ModeSwarm = jsonAttrFrom(state.ModeSwarm, postgres.ModeSwarm)
	state.LabelsSwarm = jsonAttrFrom(state.LabelsSwarm, postgres.LabelsSwarm)
	state.EndpointSpecSwarm = jsonAttrFrom(state.EndpointSpecSwarm, postgres.EndpointSpecSwarm)
	state.NetworkSwarm = jsonAttrFrom(state.NetworkSwarm, postgres.NetworkSwarm)
	state.UlimitsSwarm = jsonAttrFrom(state.UlimitsSwarm, postgres.UlimitsSwarm)

	if deployOnCreate.IsNull() || deployOnCreate.IsUnknown() {
		state.DeployOnCreate = types.BoolValue(false)
	} else {
		state.DeployOnCreate = deployOnCreate
	}
	if deployOnUpdate.IsNull() || deployOnUpdate.IsUnknown() {
		state.DeployOnUpdate = types.BoolValue(true)
	} else {
		state.DeployOnUpdate = deployOnUpdate
	}
}

func stringOrNull(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func stringList(ctx context.Context, list types.List) ([]string, bool, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, false, nil
	}
	var out []string
	diags := list.ElementsAs(ctx, &out, false)
	if diags.HasError() {
		return nil, false, diags
	}
	return out, true, nil
}

func stringListOrNull(ctx context.Context, values []string) types.List {
	if len(values) == 0 {
		return types.ListNull(types.StringType)
	}
	list, diags := types.ListValueFrom(ctx, types.StringType, values)
	if diags.HasError() {
		return types.ListNull(types.StringType)
	}
	return list
}

func parseJSONObject(raw string) (map[string]interface{}, error) {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, fmt.Errorf("expected a JSON object: %w", err)
	}
	return parsed, nil
}

func parseJSONObjectList(raw string) ([]map[string]interface{}, error) {
	var parsed []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, fmt.Errorf("expected a JSON array of objects: %w", err)
	}
	return parsed, nil
}

func jsonAttrFrom(current types.String, raw interface{}) types.String {
	if raw == nil {
		if current.IsNull() || current.IsUnknown() {
			return types.StringNull()
		}
		return current
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return current
	}
	next := string(encoded)
	if current.IsNull() || current.IsUnknown() || current.ValueString() == "" {
		return types.StringValue(next)
	}
	if jsonSemanticallyEqual(current.ValueString(), next) {
		return current
	}
	return types.StringValue(next)
}

func jsonSemanticallyEqual(left, right string) bool {
	var a, b interface{}
	if json.Unmarshal([]byte(left), &a) != nil || json.Unmarshal([]byte(right), &b) != nil {
		return false
	}
	ab, errA := json.Marshal(a)
	bb, errB := json.Marshal(b)
	return errA == nil && errB == nil && string(ab) == string(bb)
}
