package sdk_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	sdk "github.com/punk-one/iot-rule-plugin-sdk"
)

type engineRPCStub struct {
	state       map[string][]byte
	counter     string
	observation string
	health      []sdk.ReportHealthArgs
	emitted     []sdk.Event
	acked       []string
	publishAcks []sdk.AckMessage
	batches     [][]sdk.Event
	logLevel    sdk.LogLevel
	logMessages []string
	logFields   []map[string]interface{}
}

func newEngineRPCStub() *engineRPCStub {
	return &engineRPCStub{state: make(map[string][]byte)}
}

func (s *engineRPCStub) Emit(e sdk.Event) error {
	s.emitted = append(s.emitted, e)
	return nil
}

func (s *engineRPCStub) EmitWithTargets(e sdk.Event, toNodeIDs []string) error {
	s.emitted = append(s.emitted, e)
	return nil
}

func (s *engineRPCStub) Log(level sdk.LogLevel, msg string, fields map[string]interface{}) {}

func (s *engineRPCStub) LogBatch(level sdk.LogLevel, messages []string, fields []map[string]interface{}) {
	s.logLevel = level
	s.logMessages = append([]string(nil), messages...)
	s.logFields = append([]map[string]interface{}(nil), fields...)
}

func (s *engineRPCStub) IncCounter(name string, labels map[string]string) { s.counter = name }

func (s *engineRPCStub) Observe(name string, value float64, labels map[string]string) {
	s.observation = name
}

func (s *engineRPCStub) ReportHealth(args sdk.ReportHealthArgs) error {
	s.health = append(s.health, args)
	return nil
}

func (s *engineRPCStub) Ack(eventID string) error {
	s.acked = append(s.acked, eventID)
	return nil
}

func (s *engineRPCStub) PublishAck(ack sdk.AckMessage) error {
	s.publishAcks = append(s.publishAcks, ack)
	return nil
}

func (s *engineRPCStub) EmitBatch(events []sdk.Event) error {
	copied := append([]sdk.Event(nil), events...)
	s.batches = append(s.batches, copied)
	return nil
}

func (s *engineRPCStub) GetState(key string) ([]byte, error) { return s.state[key], nil }

func (s *engineRPCStub) SetState(key string, value []byte) error {
	s.state[key] = append([]byte(nil), value...)
	return nil
}

func (s *engineRPCStub) SetStateWithTTL(key string, value []byte, ttl time.Duration) error {
	return s.SetState(key, value)
}

func (s *engineRPCStub) DeleteState(key string) error {
	delete(s.state, key)
	return nil
}

func (s *engineRPCStub) ExecuteConnector(req sdk.ConnectorRequest) (sdk.ConnectorResponse, error) {
	return sdk.ConnectorResponse{}, nil
}

func (s *engineRPCStub) CurrentResourceStatus(resourceRef string) (sdk.ResourceStatusEvent, bool) {
	return sdk.ResourceStatusEvent{}, false
}

func (s *engineRPCStub) NextResourceEvent(timeout time.Duration) (sdk.ResourceStatusEvent, bool, error) {
	return sdk.ResourceStatusEvent{}, false, nil
}

func (s *engineRPCStub) OpenConnectorStream(req sdk.StreamOpenRequest) (sdk.StreamOpenResponse, error) {
	return sdk.StreamOpenResponse{}, nil
}

func (s *engineRPCStub) ReceiveConnectorStream(req sdk.StreamReceiveRequest) (sdk.StreamReceiveResponse, error) {
	return sdk.StreamReceiveResponse{}, nil
}

func (s *engineRPCStub) AckConnectorStream(req sdk.StreamAckRequest) error { return nil }

func (s *engineRPCStub) NackConnectorStream(req sdk.StreamNackRequest) error { return nil }

func (s *engineRPCStub) GrantConnectorStream(req sdk.StreamGrantCreditsRequest) error {
	return nil
}

func (s *engineRPCStub) CloseConnectorStream(req sdk.StreamCloseRequest) error { return nil }


type stateStoreStub struct {
	state map[string][]byte
}

func newStateStoreStub() *stateStoreStub {
	return &stateStoreStub{state: make(map[string][]byte)}
}

func (s *stateStoreStub) GetState(ctx context.Context, key string, state interface{}) error {
	data, ok := s.state[key]
	if !ok {
		return nil
	}
	return json.Unmarshal(data, state)
}

func (s *stateStoreStub) SetState(ctx context.Context, key string, state interface{}) error {
	return s.SetStateWithTTL(ctx, key, state, 0)
}

func (s *stateStoreStub) SetStateWithTTL(ctx context.Context, key string, state interface{}, ttl time.Duration) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	s.state[key] = data
	return nil
}

func (s *stateStoreStub) DeleteState(ctx context.Context, key string) error {
	delete(s.state, key)
	return nil
}

func TestNewStateManagerUsesStableStateStoreInterface(t *testing.T) {
	manager := sdk.NewStateManager("rule-test", "node-test", newStateStoreStub())

	type snapshot struct {
		Value int `json:"value"`
	}

	key := manager.GenerateKey([]string{"device-1"})
	if err := manager.SetState(key, &snapshot{Value: 42}); err != nil {
		t.Fatalf("SetState failed: %v", err)
	}

	var got snapshot
	if err := manager.GetState(key, &got); err != nil {
		t.Fatalf("GetState failed: %v", err)
	}
	if got.Value != 42 {
		t.Fatalf("expected value 42, got %d", got.Value)
	}
}

func TestStableRuntimeHelpersForwardToEngineRPC(t *testing.T) {
	engine := newEngineRPCStub()

	logger := sdk.NewRPCLoggerFromEngine(engine, "rule-test", "node-test")
	logger.LogBatch(sdk.LogLevelInfo, []string{"processed"}, []map[string]interface{}{{"count": 1}})
	if engine.logLevel != sdk.LogLevelInfo {
		t.Fatalf("expected info log level, got %s", engine.logLevel)
	}
	if len(engine.logMessages) != 1 || engine.logMessages[0] != "processed" {
		t.Fatalf("expected log batch to be forwarded, got %#v", engine.logMessages)
	}

	emitter := sdk.NewRPCEmitterFromEngine(engine)
	event := sdk.NewEvent(map[string]interface{}{"value": 1}, nil)
	if err := emitter.Publish(event); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}
	if err := emitter.EmitTo("branch-a", event); err != nil {
		t.Fatalf("EmitTo failed: %v", err)
	}
	if err := emitter.Ack("evt-1"); err != nil {
		t.Fatalf("Ack failed: %v", err)
	}
	ack := sdk.AckMessage{EventID: "evt-1", Status: sdk.AckStatusSuccess}
	if err := emitter.PublishAck(ack); err != nil {
		t.Fatalf("PublishAck failed: %v", err)
	}
	if err := emitter.EmitBatch([]sdk.Event{event}); err != nil {
		t.Fatalf("EmitBatch failed: %v", err)
	}

	if len(engine.emitted) != 2 {
		t.Fatalf("expected 2 emitted events, got %d", len(engine.emitted))
	}
	if got := engine.emitted[1].Metadata["route_label"]; got != "branch-a" {
		t.Fatalf("expected routed event label branch-a, got %q", got)
	}
	if len(engine.acked) != 1 || engine.acked[0] != "evt-1" {
		t.Fatalf("expected ack to be forwarded, got %#v", engine.acked)
	}
	if len(engine.publishAcks) != 1 || engine.publishAcks[0].EventID != "evt-1" {
		t.Fatalf("expected publish ack to be forwarded, got %#v", engine.publishAcks)
	}
	if len(engine.batches) != 1 || len(engine.batches[0]) != 1 {
		t.Fatalf("expected one emitted batch, got %#v", engine.batches)
	}

	metrics := sdk.NewRPCMetricsFromEngine(engine, "rule-test", "node-test")
	metrics.IncCounter("events_processed", map[string]string{"node": "node-test"})
	metrics.Observe("latency_ms", 12, nil)
	if engine.counter != "events_processed" {
		t.Fatalf("expected metric counter to be forwarded, got %q", engine.counter)
	}
	if engine.observation != "latency_ms" {
		t.Fatalf("expected observation to be forwarded, got %q", engine.observation)
	}
}
