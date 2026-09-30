package sdk_test

import (
	"testing"

	sdk "github.com/punk-one/iot-rule-plugin-sdk"
)

func TestDefaultHandshakeConfigIsImmutableV1Copy(t *testing.T) {
	original := sdk.HandshakeConfig
	t.Cleanup(func() { sdk.HandshakeConfig = original })
	sdk.HandshakeConfig.ProtocolVersion = 99
	sdk.HandshakeConfig.MagicCookieKey = "mutated"

	got := sdk.DefaultHandshakeConfig()
	if got.ProtocolVersion != 1 || got.MagicCookieKey != "PUNK_PLUGIN_MAGIC" || got.MagicCookieValue != "punk-rule-engine" {
		t.Fatalf("unexpected v1 handshake: %#v", got)
	}
	got.ProtocolVersion = 77
	if again := sdk.DefaultHandshakeConfig(); again.ProtocolVersion != 1 {
		t.Fatalf("handshake function did not return an independent copy: %#v", again)
	}
}
