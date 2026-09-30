package transport

import "github.com/punk-one/iot-rule-plugin-sdk/internal/core"

func mergeHealthOptions(primary, fallback core.HealthOptions) core.HealthOptions {
	if primary.HeartbeatInterval <= 0 {
		primary.HeartbeatInterval = fallback.HeartbeatInterval
	}
	if primary.MaxSilencePeriod <= 0 {
		primary.MaxSilencePeriod = fallback.MaxSilencePeriod
	}
	if primary.QueueCapacity <= 0 {
		primary.QueueCapacity = fallback.QueueCapacity
	}
	return primary.Normalize()
}
