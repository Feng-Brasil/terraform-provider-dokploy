package provider

import (
	"github.com/feng-brasil/terraform-provider-dokploy/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NetworkIPAMConfigModel struct {
	Subnet  types.String `tfsdk:"subnet"`
	Gateway types.String `tfsdk:"gateway"`
	IPRange types.String `tfsdk:"ip_range"`
}

type NetworkIPAMModel struct {
	Driver types.String             `tfsdk:"driver"`
	Config []NetworkIPAMConfigModel `tfsdk:"config"`
}

func nullableString(v string) types.String {
	if v == "" {
		return types.StringNull()
	}

	return types.StringValue(v)
}

func flattenNetworkIPAMModel(ipam *client.NetworkIPAM) *NetworkIPAMModel {
	if ipam == nil {
		return nil
	}

	if ipam.Driver == "" && len(ipam.Config) == 0 {
		return nil
	}

	model := &NetworkIPAMModel{
		Driver: nullableString(ipam.Driver),
	}

	for _, cfg := range ipam.Config {
		model.Config = append(model.Config, NetworkIPAMConfigModel{
			Subnet:  nullableString(cfg.Subnet),
			Gateway: nullableString(cfg.Gateway),
			IPRange: nullableString(cfg.IPRange),
		})
	}

	return model
}

func expandNetworkIPAMModel(model *NetworkIPAMModel) *client.NetworkIPAM {
	if model == nil {
		return nil
	}

	ipam := &client.NetworkIPAM{}

	if !model.Driver.IsNull() && !model.Driver.IsUnknown() {
		ipam.Driver = model.Driver.ValueString()
	}

	for _, cfg := range model.Config {
		entry := client.NetworkIPAMConfig{}
		hasContent := false

		if !cfg.Subnet.IsNull() && !cfg.Subnet.IsUnknown() {
			entry.Subnet = cfg.Subnet.ValueString()
			hasContent = true
		}
		if !cfg.Gateway.IsNull() && !cfg.Gateway.IsUnknown() {
			entry.Gateway = cfg.Gateway.ValueString()
			hasContent = true
		}
		if !cfg.IPRange.IsNull() && !cfg.IPRange.IsUnknown() {
			entry.IPRange = cfg.IPRange.ValueString()
			hasContent = true
		}

		if hasContent {
			ipam.Config = append(ipam.Config, entry)
		}
	}

	if ipam.Driver == "" && len(ipam.Config) == 0 {
		return nil
	}

	return ipam
}
