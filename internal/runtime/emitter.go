package runtime

import (
	"fmt"

	"github.com/punk-one/iot-rule-plugin-sdk/internal/core"
)

type EngineRPCEmitter struct {
	engine EngineBridge
}

func NewRPCEmitterFromEngine(engine EngineBridge) core.Emitter {
	return &EngineRPCEmitter{engine: engine}
}

func (e *EngineRPCEmitter) Publish(evt core.Event) error { return e.engine.Emit(evt) }
func (e *EngineRPCEmitter) Ack(eventID string) error     { return e.engine.Ack(eventID) }

func (e *EngineRPCEmitter) PublishAck(ack core.AckMessage) error {
	return e.engine.PublishAck(ack)
}

func (e *EngineRPCEmitter) EmitTo(label string, evt core.Event) error {
	if evt.Metadata == nil {
		evt.Metadata = make(map[string]string)
	}
	evt.Metadata["route_label"] = label
	return e.engine.Emit(evt)
}

func (e *EngineRPCEmitter) EmitBatch(events []core.Event) error {
	return e.engine.EmitBatch(events)
}

type NoOpEmitter struct{}

func (e *NoOpEmitter) Publish(evt core.Event) error { return fmt.Errorf("emitter not available") }
func (e *NoOpEmitter) EmitTo(label string, evt core.Event) error {
	return fmt.Errorf("emitter not available")
}
func (e *NoOpEmitter) Ack(eventID string) error             { return nil }
func (e *NoOpEmitter) PublishAck(ack core.AckMessage) error { return nil }
func (e *NoOpEmitter) EmitBatch(events []core.Event) error {
	return fmt.Errorf("emitter not available")
}
