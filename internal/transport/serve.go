package transport

import (
	"github.com/hashicorp/go-plugin"
	"github.com/punk-one/iot-rule-plugin-sdk/internal/core"
)

func Serve(impl core.Plugin, options core.HealthOptions) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: DefaultHandshakeConfig(),
		Plugins: map[string]plugin.Plugin{
			"plugin": &PluginRPC{
				Impl:          impl,
				defaultHealth: options,
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

// ServeFactory serves one operating-system process while creating an isolated
// Plugin object for every host-side Dispense call. This is the recommended mode
// for engines that share one plugin binary process across multiple rule nodes.
func ServeFactory(factory func() core.Plugin, options core.HealthOptions) {
	rpcPlugin := &PluginRPC{
		Factory:       factory,
		defaultHealth: options,
	}
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: DefaultHandshakeConfig(),
		Plugins: map[string]plugin.Plugin{
			"plugin":         rpcPlugin,
			"plugin_factory": rpcPlugin,
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

func ServeConnector(impl core.ConnectorPlugin) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: DefaultHandshakeConfig(),
		Plugins: map[string]plugin.Plugin{
			"connector": &ConnectorRPC{Impl: impl},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
