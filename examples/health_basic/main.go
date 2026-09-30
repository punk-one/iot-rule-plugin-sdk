package main

import (
	"time"

	sdk "github.com/punk-one/iot-rule-plugin-sdk"
)

type healthAwareSource struct {
	sdk.BasePlugin
	ctx sdk.RuntimeContext
}

func (s *healthAwareSource) Info() sdk.PluginInfo {
	return sdk.PluginInfo{
		ID:      "example-health-source",
		Name:    "Example Health Source",
		Version: "1.8.1",
		Type:    sdk.PluginTypeSource,
		Capabilities: sdk.PluginCapabilities{
			Stateful: false,
		},
	}
}

func (s *healthAwareSource) Start(ctx sdk.RuntimeContext) error {
	s.ctx = ctx

	if err := s.connectToUpstream(); err != nil {
		_ = ctx.Health().Unhealthy("NET_CONNECT_FAILED", err.Error(), map[string]string{
			"component": "plc",
			"mode":      "bootstrap",
		})
		return sdk.RetryableError(err)
	}

	_ = ctx.Health().Healthy("upstream connected", map[string]string{
		"component": "plc",
		"mode":      "streaming",
	})

	go s.loop()
	return nil
}

func (s *healthAwareSource) loop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		event := sdk.NewEvent(map[string]interface{}{
			"value":  42,
			"source": "health-basic",
		}, map[string]string{
			"device_id": "plc-1",
		})

		if err := s.ctx.Emitter().Publish(event); err != nil {
			_ = s.ctx.Health().Degraded("NET_EMIT_SLOW", err.Error(), map[string]string{
				"component": "emitter",
			})
			continue
		}

		_ = s.ctx.Health().Heartbeat("source active", map[string]string{
			"component": "plc",
		})
	}
}

func (s *healthAwareSource) connectToUpstream() error {
	return nil
}

func main() {
	sdk.ServeFactory(func() sdk.Plugin { return &healthAwareSource{} }, sdk.ServeOptions{
		Health: sdk.HealthOptions{
			HeartbeatInterval: 10 * time.Second,
			MaxSilencePeriod:  time.Minute,
			QueueCapacity:     64,
		},
	})
}
