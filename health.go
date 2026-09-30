package sdk

import (
	"time"

	"github.com/punk-one/iot-rule-plugin-sdk/internal/core"
	healthruntime "github.com/punk-one/iot-rule-plugin-sdk/internal/healthruntime"
)

type HealthStatus = core.HealthStatus

const (
	HealthUnknown      = core.HealthUnknown
	HealthInitializing = core.HealthInitializing
	HealthHealthy      = core.HealthHealthy
	HealthDegraded     = core.HealthDegraded
	HealthRecovering   = core.HealthRecovering
	HealthUnhealthy    = core.HealthUnhealthy
	HealthStopped      = core.HealthStopped
)

type HealthKind = core.HealthKind

const (
	HealthKindState     = core.HealthKindState
	HealthKindHeartbeat = core.HealthKindHeartbeat
)

type HealthOptions = core.HealthOptions

func DefaultHealthOptions() HealthOptions {
	return core.DefaultHealthOptions()
}

type HealthReport = core.HealthReport
type HealthHeartbeat = core.HealthHeartbeat
type ReportHealthArgs = core.ReportHealthArgs
type HealthReporter = core.HealthReporter
type HealthRuntimeStats = healthruntime.Stats

// FlushHealthReporter waits until currently accepted health messages have
// been sent or the timeout expires. Unbuffered reporters are a no-op.
func FlushHealthReporter(reporter HealthReporter, timeout time.Duration) error {
	return healthruntime.Flush(reporter, timeout)
}

// GetHealthRuntimeStats returns health coalescing and sender error counters.
func GetHealthRuntimeStats(reporter HealthReporter) HealthRuntimeStats {
	return healthruntime.Snapshot(reporter)
}

func newNoopHealthReporter() HealthReporter {
	return healthruntime.NewNoop()
}
