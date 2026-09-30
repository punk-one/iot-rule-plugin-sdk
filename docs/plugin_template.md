# v1 插件模板

下面的骨架适用于 Source、Processor 和 Sink。插件只依赖公共 SDK，不应导入
`iot-rule-engine/internal`。

```go
package main

import (
	"time"

	sdk "github.com/punk-one/iot-rule-plugin-sdk"
)

type Plugin struct {
	sdk.BasePlugin
	ctx sdk.RuntimeContext
}

func (p *Plugin) Info() sdk.PluginInfo {
	return sdk.PluginInfo{
		ID:      "example-processor",
		Name:    "Example Processor",
		Version: "1.0.0",
		Type:    sdk.PluginTypeProcessor,
	}
}

func (p *Plugin) Start(ctx sdk.RuntimeContext) error {
	p.ctx = ctx
	return ctx.Health().Healthy("started", nil)
}

func (p *Plugin) OnEvent(event sdk.Event) error {
	// 当前调用独占 event；发往多个异步目标时，每个目标都应使用 Clone。
	owned := event.Clone()
	owned.Payload["processed"] = true
	return p.ctx.Emitter().Publish(owned)
}

func (p *Plugin) Stop() error {
	if p.ctx == nil {
		return nil
	}
	_ = p.ctx.Health().Report(sdk.HealthReport{Status: sdk.HealthStopped})
	return sdk.FlushHealthReporter(p.ctx.Health(), 3*time.Second)
}

func main() {
	sdk.ServeFactory(func() sdk.Plugin { return &Plugin{} }, sdk.ServeOptions{})
}
```

若插件使用 Connector，应声明 `PluginCapabilities.ConnectorBinding`；没有绑定能力时 Connector 调用会返回
`sdk.ErrNotImplemented`，不能把空响应视为成功。

factory 每次必须返回新对象，使不同规则节点不共享配置和运行状态。

v1 约束：不要实现未声明的内部接口，不要修改 `sdk.HandshakeConfig`，不要解析
Event ID 中的时间值；只把它当作以 `evt_` 开头的不透明唯一标识。
