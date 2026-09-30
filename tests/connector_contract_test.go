package sdk_test

import (
	"encoding/json"
	"testing"

	sdk "github.com/punk-one/iot-rule-plugin-sdk"
)

func TestPluginCapabilitiesConnectorContractJSON(t *testing.T) {
	info := sdk.PluginInfo{
		ID:      "source-s7",
		Name:    "S7 Source",
		Version: "1.7.3",
		Type:    sdk.PluginTypeSource,
		Capabilities: sdk.PluginCapabilities{
			ConfigSchema: `{"type":"object"}`,
			ConnectorBinding: &sdk.ConnectorBindingSpec{
				Required:             true,
				AcceptedFamilies:     []string{"s7"},
				RequiredCapabilities: []string{"read"},
			},
		},
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal plugin info failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal plugin info failed: %v", err)
	}

	capabilities, ok := decoded["capabilities"].(map[string]interface{})
	if !ok {
		t.Fatalf("capabilities missing: %s", string(data))
	}
	binding, ok := capabilities["connector_binding"].(map[string]interface{})
	if !ok {
		t.Fatalf("connector_binding missing: %s", string(data))
	}
	if binding["required"] != true {
		t.Fatalf("expected required=true, got %#v", binding["required"])
	}
}

func TestProcessorPluginCapabilitiesConnectorContractJSON(t *testing.T) {
	info := sdk.PluginInfo{
		ID:      "processor-enricher",
		Name:    "Processor Enricher",
		Version: "1.7.3",
		Type:    sdk.PluginTypeProcessor,
		Capabilities: sdk.PluginCapabilities{
			ConfigSchema: `{"type":"object"}`,
			ConnectorBinding: &sdk.ConnectorBindingSpec{
				Required:             false,
				AcceptedFamilies:     []string{"http", "rest"},
				RequiredCapabilities: []string{"read"},
			},
		},
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal plugin info failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal plugin info failed: %v", err)
	}

	if decoded["type"] != string(sdk.PluginTypeProcessor) {
		t.Fatalf("expected processor type, got %#v", decoded["type"])
	}

	capabilities, ok := decoded["capabilities"].(map[string]interface{})
	if !ok {
		t.Fatalf("capabilities missing: %s", string(data))
	}
	binding, ok := capabilities["connector_binding"].(map[string]interface{})
	if !ok {
		t.Fatalf("connector_binding missing: %s", string(data))
	}
	if binding["required"] != false {
		t.Fatalf("expected required=false, got %#v", binding["required"])
	}

	families, ok := binding["accepted_families"].([]interface{})
	if !ok || len(families) != 2 {
		t.Fatalf("unexpected accepted_families: %#v", binding["accepted_families"])
	}
	if families[0] != "http" || families[1] != "rest" {
		t.Fatalf("unexpected accepted_families order/content: %#v", families)
	}
}

func TestConnectorDescriptorExposedOnPluginCapabilities(t *testing.T) {
	caps := sdk.PluginCapabilities{
		ConnectorDescriptor: &sdk.ConnectorDescriptor{
			Family:        "timescaledb",
			Label:         "TimescaleDB / PostgreSQL",
			Capabilities:  []string{"read", "write", "connection_pool"},
			DisplayFields: []string{"host", "port", "database"},
		},
	}

	if caps.ConnectorDescriptor == nil {
		t.Fatal("expected connector descriptor to be set")
	}
	if caps.ConnectorDescriptor.Family != "timescaledb" {
		t.Fatalf("unexpected family: %s", caps.ConnectorDescriptor.Family)
	}
	if len(caps.ConnectorDescriptor.DisplayFields) != 3 {
		t.Fatalf("unexpected display fields: %#v", caps.ConnectorDescriptor.DisplayFields)
	}
}
