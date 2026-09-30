package sdk

import (
	"github.com/punk-one/iot-rule-plugin-sdk/internal/core"
	transport "github.com/punk-one/iot-rule-plugin-sdk/internal/transport"
)

// ServeOptions 控制 sdk.Serve 的运行行为。
type ServeOptions struct {
	Health HealthOptions
}

// PluginFactory creates one isolated plugin object for one engine rule-node
// instance. Implementations must return a new object on every call.
type PluginFactory func() Plugin

// Serve 以 SDK 约定的默认配置启动插件进程。
// Deprecated: 多规则引擎应使用 ServeFactory，避免多个规则共享可变插件对象。
func Serve(impl Plugin, options ServeOptions) {
	transport.Serve(impl, options.Health)
}

// ServeFactory 启动一个插件进程，并为每个规则节点创建独立插件对象。
// 同一进程中的不同对象不会共享配置、上下文和生命周期状态。
func ServeFactory(factory PluginFactory, options ServeOptions) {
	transport.ServeFactory(func() core.Plugin { return factory() }, options.Health)
}

// ServeConnector 以 SDK 约定的默认配置启动 connector 插件进程。
func ServeConnector(impl ConnectorPlugin) {
	transport.ServeConnector(impl)
}
