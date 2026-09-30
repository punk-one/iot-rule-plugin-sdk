package main

import sdk "github.com/punk-one/iot-rule-plugin-sdk"

type processor struct {
	sdk.BasePlugin
	ctx sdk.RuntimeContext
}

func (p *processor) Info() sdk.PluginInfo {
	return sdk.PluginInfo{
		ID:      "example-processor",
		Name:    "Example Processor",
		Version: "1.8.1",
		Type:    sdk.PluginTypeProcessor,
		Capabilities: sdk.PluginCapabilities{
			SupportBatch: true,
		},
	}
}

func (p *processor) Start(ctx sdk.RuntimeContext) error {
	p.ctx = ctx
	_ = ctx.Health().Healthy("processor started", map[string]string{"component": "processor"})
	return nil
}

func (p *processor) OnEvent(event sdk.Event) error {
	event.Payload["processed"] = true
	return p.ctx.Emitter().Publish(event)
}

func (p *processor) OnEvents(events []sdk.Event) error {
	for index := range events {
		events[index].Payload["processed"] = true
	}
	return p.ctx.Emitter().EmitBatch(events)
}

func main() {
	sdk.ServeFactory(func() sdk.Plugin { return &processor{} }, sdk.ServeOptions{})
}
