package runtime

import (
	"github.com/punk-one/iot-rule-plugin-sdk/internal/core"
	healthruntime "github.com/punk-one/iot-rule-plugin-sdk/internal/healthruntime"
)

type engineHealthSender struct {
	engine EngineBridge
}

func NewEngineHealthReporter(engine EngineBridge, options core.HealthOptions) core.HealthReporter {
	if engine == nil {
		return healthruntime.NewNoop()
	}
	return healthruntime.NewAsync(&engineHealthSender{engine: engine}, options)
}

func (s *engineHealthSender) SendHealth(args core.ReportHealthArgs) error {
	if s.engine == nil {
		return nil
	}
	return s.engine.ReportHealth(args)
}
