package transport

import "github.com/hashicorp/go-plugin"

const (
	protocolVersion  = 1
	magicCookieKey   = "PUNK_PLUGIN_MAGIC"
	magicCookieValue = "punk-rule-engine"
)

// DefaultHandshakeConfig returns a fresh copy of the immutable v1 handshake.
func DefaultHandshakeConfig() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		ProtocolVersion:  protocolVersion,
		MagicCookieKey:   magicCookieKey,
		MagicCookieValue: magicCookieValue,
	}
}

// HandshakeConfig is retained for source compatibility. Internal transport
// paths deliberately use DefaultHandshakeConfig so mutations cannot alter the
// v1 wire contract.
var HandshakeConfig = DefaultHandshakeConfig()
