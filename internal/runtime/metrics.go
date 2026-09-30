package runtime

import "github.com/punk-one/iot-rule-plugin-sdk/internal/core"

type EngineRPCMetrics struct {
	engine EngineBridge
}

type DefaultMetrics struct{}

func (m *DefaultMetrics) IncCounter(name string, labels map[string]string)             {}
func (m *DefaultMetrics) Observe(name string, value float64, labels map[string]string) {}

func NewRPCMetricsFromEngine(engine EngineBridge, ruleID, nodeID string) core.Metrics {
	if engine == nil {
		return &DefaultMetrics{}
	}
	return &EngineRPCMetrics{engine: engine}
}

func (m *EngineRPCMetrics) IncCounter(name string, labels map[string]string) {
	if m.engine == nil {
		return
	}
	m.engine.IncCounter(name, labels)
}

func (m *EngineRPCMetrics) Observe(name string, value float64, labels map[string]string) {
	if m.engine == nil {
		return
	}
	m.engine.Observe(name, value, labels)
}
