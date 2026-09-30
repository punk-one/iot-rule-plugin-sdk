package main

import sdk "github.com/punk-one/iot-rule-plugin-sdk"

type source struct {
	sdk.BasePlugin
	ctx sdk.RuntimeContext
}

func (s *source) Info() sdk.PluginInfo {
	return sdk.PluginInfo{
		ID:      "example-source",
		Name:    "Example Source",
		Version: "1.8.1",
		Type:    sdk.PluginTypeSource,
	}
}

func (s *source) Start(ctx sdk.RuntimeContext) error {
	s.ctx = ctx
	_ = ctx.Health().Healthy("source started", nil)
	event := sdk.NewEvent(map[string]interface{}{"data": "value"}, nil)
	return s.ctx.Emitter().Publish(event)
}

func main() {
	sdk.ServeFactory(func() sdk.Plugin { return &source{} }, sdk.ServeOptions{})
}
