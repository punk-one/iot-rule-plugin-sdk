package sdk_test

import (
	"testing"

	sdk "github.com/punk-one/iot-rule-plugin-sdk"
)

func TestConnectorSessionPolicyConstants(t *testing.T) {
	want := map[string]string{
		"per_operation":   sdk.ConnectorSessionPolicyPerOperation,
		"persistent":      sdk.ConnectorSessionPolicyPersistent,
		"pooled":          sdk.ConnectorSessionPolicyPooled,
		"stateless":       sdk.ConnectorSessionPolicyStateless,
		"idle_timeout":    sdk.ConnectorSessionPolicyIdleTimeout,
		"scheduled_reuse": sdk.ConnectorSessionPolicyScheduledReuse,
	}
	for expected, actual := range want {
		if actual != expected {
			t.Fatalf("session policy constant changed: want %q, got %q", expected, actual)
		}
	}
}

func TestConnectorResourcePolicyHelpersRoundTrip(t *testing.T) {
	resource := sdk.ConnectorResource{
		ID:       "resource-1",
		PluginID: "connect-dummy",
	}

	health := sdk.HealthPolicy{
		Enabled:    true,
		IntervalMS: 1500,
		TimeoutMS:  500,
		Target:     "health",
		Priority:   "high",
	}
	if err := resource.SetHealthPolicy(health); err != nil {
		t.Fatalf("SetHealthPolicy failed: %v", err)
	}

	quota := sdk.QuotaPolicy{
		MaxInflightRequests: 32,
		MaxInflightWeight:   64,
		DefaultUsageWeight:  2,
	}
	if err := resource.SetQuotaPolicy(quota); err != nil {
		t.Fatalf("SetQuotaPolicy failed: %v", err)
	}

	gotHealth, err := resource.GetHealthPolicy()
	if err != nil {
		t.Fatalf("GetHealthPolicy failed: %v", err)
	}
	if gotHealth != health {
		t.Fatalf("unexpected health policy: %#v", gotHealth)
	}

	gotQuota, err := resource.GetQuotaPolicy()
	if err != nil {
		t.Fatalf("GetQuotaPolicy failed: %v", err)
	}
	if gotQuota != quota {
		t.Fatalf("unexpected quota policy: %#v", gotQuota)
	}
}

func TestConnectorResourcePolicyHelpersClearZeroValues(t *testing.T) {
	resource := sdk.ConnectorResource{
		ID:           "resource-2",
		PluginID:     "connect-dummy",
		HealthPolicy: []byte(`{"enabled":true}`),
		QuotaPolicy:  []byte(`{"max_inflight_requests":2}`),
	}

	if err := resource.SetHealthPolicy(sdk.HealthPolicy{}); err != nil {
		t.Fatalf("SetHealthPolicy zero failed: %v", err)
	}
	if err := resource.SetQuotaPolicy(sdk.QuotaPolicy{}); err != nil {
		t.Fatalf("SetQuotaPolicy zero failed: %v", err)
	}

	if len(resource.HealthPolicy) != 0 {
		t.Fatalf("expected health policy to be cleared, got %s", string(resource.HealthPolicy))
	}
	if len(resource.QuotaPolicy) != 0 {
		t.Fatalf("expected quota policy to be cleared, got %s", string(resource.QuotaPolicy))
	}
}

func TestConnectorResourcePolicyHelpersRejectInvalidJSON(t *testing.T) {
	resource := sdk.ConnectorResource{
		ID:           "resource-3",
		PluginID:     "connect-dummy",
		HealthPolicy: []byte(`{`),
		QuotaPolicy:  []byte(`{`),
	}

	if _, err := resource.GetHealthPolicy(); err == nil {
		t.Fatal("expected invalid health policy json to fail")
	}
	if _, err := resource.GetQuotaPolicy(); err == nil {
		t.Fatal("expected invalid quota policy json to fail")
	}
}
