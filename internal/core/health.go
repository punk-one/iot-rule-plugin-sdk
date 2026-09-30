package core

import healthruntime "github.com/punk-one/iot-rule-plugin-sdk/internal/healthruntime"

type HealthStatus = healthruntime.Status

const (
	HealthUnknown      = healthruntime.StatusUnknown
	HealthInitializing = healthruntime.StatusInitializing
	HealthHealthy      = healthruntime.StatusHealthy
	HealthDegraded     = healthruntime.StatusDegraded
	HealthRecovering   = healthruntime.StatusRecovering
	HealthUnhealthy    = healthruntime.StatusUnhealthy
	HealthStopped      = healthruntime.StatusStopped
)

type HealthKind = healthruntime.Kind

const (
	HealthKindState     = healthruntime.KindState
	HealthKindHeartbeat = healthruntime.KindHeartbeat
)

type HealthOptions = healthruntime.Options

func DefaultHealthOptions() HealthOptions {
	return healthruntime.DefaultOptions()
}

type HealthReport = healthruntime.Report
type HealthHeartbeat = healthruntime.Heartbeat
type ReportHealthArgs = healthruntime.ReportArgs
type HealthReporter = healthruntime.Reporter
