package runtime

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/punk-one/iot-rule-plugin-sdk/internal/core"
)

type engineBridgeStub struct {
	state           map[string][]byte
	counter         string
	observation     string
	counters        []string
	observations    []string
	health          []core.ReportHealthArgs
	streamOpenReq   core.StreamOpenRequest
	streamRecvReqs  []core.StreamReceiveRequest
	streamAckReq    core.StreamAckRequest
	streamNackReq   core.StreamNackRequest
	streamGrantReq  core.StreamGrantCreditsRequest
	streamCloseReq  core.StreamCloseRequest
	streamRecvCalls int
	nextEvent       func(timeout time.Duration) (core.ResourceStatusEvent, bool, error)
	nextEventCalls  int64
	mu              sync.Mutex
	logMessages     []string
	logFields       []map[string]interface{}
}

func newEngineBridgeStub() *engineBridgeStub {
	return &engineBridgeStub{state: make(map[string][]byte)}
}

func (s *engineBridgeStub) Emit(e core.Event) error                                { return nil }
func (s *engineBridgeStub) EmitWithTargets(e core.Event, toNodeIDs []string) error { return nil }
func (s *engineBridgeStub) Log(level core.LogLevel, msg string, fields map[string]interface{}) {
}
func (s *engineBridgeStub) LogBatch(level core.LogLevel, messages []string, fields []map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logMessages = append(s.logMessages, messages...)
	s.logFields = append(s.logFields, fields...)
}
func (s *engineBridgeStub) recordCounter(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters = append(s.counters, name)
}
func (s *engineBridgeStub) recordObservation(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observations = append(s.observations, name)
}
func (s *engineBridgeStub) hasCounter(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, got := range s.counters {
		if got == name {
			return true
		}
	}
	return false
}
func (s *engineBridgeStub) hasObservation(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, got := range s.observations {
		if got == name {
			return true
		}
	}
	return false
}
func (s *engineBridgeStub) nextResourceEventCallsCount() int64 {
	return atomic.LoadInt64(&s.nextEventCalls)
}
func (s *engineBridgeStub) ReportHealth(args core.ReportHealthArgs) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.health = append(s.health, args)
	return nil
}
func (s *engineBridgeStub) healthMessagesSnapshot() []core.ReportHealthArgs {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]core.ReportHealthArgs, len(s.health))
	copy(out, s.health)
	return out
}
func (s *engineBridgeStub) Ack(eventID string) error             { return nil }
func (s *engineBridgeStub) PublishAck(ack core.AckMessage) error { return nil }
func (s *engineBridgeStub) EmitBatch(events []core.Event) error  { return nil }
func (s *engineBridgeStub) GetState(key string) ([]byte, error)  { return s.state[key], nil }
func (s *engineBridgeStub) DeleteState(key string) error         { delete(s.state, key); return nil }
func (s *engineBridgeStub) SetStateWithTTL(key string, value []byte, ttl time.Duration) error {
	return s.SetState(key, value)
}
func (s *engineBridgeStub) SetState(key string, value []byte) error {
	s.state[key] = append([]byte(nil), value...)
	return nil
}
func (s *engineBridgeStub) IncCounter(name string, labels map[string]string) {
	s.counter = name
	s.recordCounter(name)
}
func (s *engineBridgeStub) Observe(name string, value float64, labels map[string]string) {
	s.observation = name
	s.recordObservation(name)
}
func (s *engineBridgeStub) ExecuteConnector(req core.ConnectorRequest) (core.ConnectorResponse, error) {
	return core.ConnectorResponse{}, nil
}
func (s *engineBridgeStub) CurrentResourceStatus(resourceRef string) (core.ResourceStatusEvent, bool) {
	return core.ResourceStatusEvent{}, false
}
func (s *engineBridgeStub) NextResourceEvent(timeout time.Duration) (core.ResourceStatusEvent, bool, error) {
	atomic.AddInt64(&s.nextEventCalls, 1)
	if s.nextEvent != nil {
		return s.nextEvent(timeout)
	}
	return core.ResourceStatusEvent{}, false, nil
}
func (s *engineBridgeStub) OpenConnectorStream(req core.StreamOpenRequest) (core.StreamOpenResponse, error) {
	s.streamOpenReq = req
	return core.StreamOpenResponse{
		StreamID:       "stream-1",
		InitialCredits: req.InitialCredits,
	}, nil
}
func (s *engineBridgeStub) ReceiveConnectorStream(req core.StreamReceiveRequest) (core.StreamReceiveResponse, error) {
	s.streamRecvCalls++
	s.streamRecvReqs = append(s.streamRecvReqs, req)

	limit := req.MaxMessages
	if limit <= 0 {
		limit = 1
	}
	messages := make([]core.StreamMessage, 0, limit)
	for i := 0; i < limit; i++ {
		messages = append(messages, core.StreamMessage{
			StreamID:            req.StreamID,
			DeliveryID:          "delivery-" + string(rune('1'+i)),
			Sequence:            uint64(i + 1),
			Topic:               "factory/device-1/status",
			RawPayload:          []byte(`{"value":42}`),
			Metadata:            map[string]string{"mqtt_topic": "factory/device-1/status"},
			PublishedAtUnixNano: 123,
		})
	}
	return core.StreamReceiveResponse{Messages: messages}, nil
}
func (s *engineBridgeStub) AckConnectorStream(req core.StreamAckRequest) error {
	s.streamAckReq = req
	return nil
}
func (s *engineBridgeStub) NackConnectorStream(req core.StreamNackRequest) error {
	s.streamNackReq = req
	return nil
}
func (s *engineBridgeStub) GrantConnectorStream(req core.StreamGrantCreditsRequest) error {
	s.streamGrantReq = req
	return nil
}
func (s *engineBridgeStub) CloseConnectorStream(req core.StreamCloseRequest) error {
	s.streamCloseReq = req
	return nil
}

func waitForHealthMessages(t *testing.T, stub *engineBridgeStub, expected int) []core.ReportHealthArgs {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		messages := stub.healthMessagesSnapshot()
		if len(messages) >= expected {
			return messages
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("expected at least %d health messages, got %d", expected, len(stub.healthMessagesSnapshot()))
	return nil
}

func TestRuntimeContextCloseFlushesAndStopsLogger(t *testing.T) {
	engine := newEngineBridgeStub()
	ctx := NewPluginRuntimeContext(engine, "rule-test", "node-test", core.DefaultHealthOptions(), nil, false)
	nested := map[string]interface{}{"values": []int{1, 2}}
	ctx.Logger().Info("before-close", core.Field{Key: "nested", Value: nested})
	nested["values"].([]int)[0] = 9

	closer, ok := ctx.(interface{ Close() error })
	if !ok {
		t.Fatal("expected runtime context to be closable")
	}
	if err := closer.Close(); err != nil {
		t.Fatalf("first close failed: %v", err)
	}
	if err := closer.Close(); err != nil {
		t.Fatalf("second close failed: %v", err)
	}

	engine.mu.Lock()
	defer engine.mu.Unlock()
	if len(engine.logMessages) != 1 || engine.logMessages[0] != "before-close" {
		t.Fatalf("expected final log batch to flush once, got %#v", engine.logMessages)
	}
	gotNested := engine.logFields[0]["nested"].(map[string]interface{})
	if gotNested["values"].([]int)[0] != 1 {
		t.Fatal("logger retained caller-owned mutable field data")
	}

	ctx.Logger().Info("after-close")
	if len(engine.logMessages) != 1 {
		t.Fatal("logger accepted a record after close")
	}
}

func TestNewPluginRuntimeContextProvidesStateAndMetrics(t *testing.T) {
	engine := newEngineBridgeStub()
	ctx := NewPluginRuntimeContext(engine, "rule-test", "node-test", core.DefaultHealthOptions(), nil, false)

	if ctx.RuleID() != "rule-test" || ctx.NodeID() != "node-test" {
		t.Fatalf("unexpected context identity: %s/%s", ctx.RuleID(), ctx.NodeID())
	}

	stateful, ok := ctx.(interface{ State() core.StateManager })
	if !ok {
		t.Fatalf("expected plugin runtime context to expose State()")
	}
	if stateful.State() == nil {
		t.Fatalf("expected non-nil state manager")
	}

	type snapshot struct {
		Value int `json:"value"`
	}

	key := stateful.State().GenerateKey([]string{"device-1"})
	if err := stateful.State().SetState(key, &snapshot{Value: 42}); err != nil {
		t.Fatalf("SetState failed: %v", err)
	}

	var got snapshot
	if err := stateful.State().GetState(key, &got); err != nil {
		t.Fatalf("GetState failed: %v", err)
	}
	if got.Value != 42 {
		t.Fatalf("expected value 42, got %d", got.Value)
	}

	ctx.Metrics().IncCounter("events_processed", map[string]string{"node": "node-test"})
	ctx.Metrics().Observe("latency_ms", 12, nil)
	if engine.counter != "events_processed" {
		t.Fatalf("expected metric counter to be forwarded, got %q", engine.counter)
	}
	if engine.observation != "latency_ms" {
		t.Fatalf("expected observation to be forwarded, got %q", engine.observation)
	}

	if closer, ok := ctx.Health().(interface{ Close() error }); ok {
		t.Cleanup(func() { _ = closer.Close() })
	}
	if err := ctx.Health().Healthy("ok", map[string]string{"phase": "test"}); err != nil {
		t.Fatalf("expected health reporter to be available: %v", err)
	}
}

func TestHealthReporterSuppressesDuplicateState(t *testing.T) {
	engine := newEngineBridgeStub()
	ctx := NewPluginRuntimeContext(engine, "rule-test", "node-test", core.HealthOptions{
		HeartbeatInterval: time.Hour,
		MaxSilencePeriod:  time.Minute,
		QueueCapacity:     8,
	}, nil, false)
	if closer, ok := ctx.Health().(interface{ Close() error }); ok {
		t.Cleanup(func() { _ = closer.Close() })
	}

	if err := ctx.Health().Healthy("connected", map[string]string{"target": "plc-1"}); err != nil {
		t.Fatalf("first report failed: %v", err)
	}
	if err := ctx.Health().Healthy("connected", map[string]string{"target": "plc-1"}); err != nil {
		t.Fatalf("second report failed: %v", err)
	}

	got := waitForHealthMessages(t, engine, 1)
	if len(got) != 1 {
		t.Fatalf("expected 1 report after suppression, got %d", len(got))
	}
	if got[0].Kind != core.HealthKindState {
		t.Fatalf("expected state report, got %s", got[0].Kind)
	}
	if got[0].Status != core.HealthHealthy {
		t.Fatalf("expected healthy status, got %s", got[0].Status)
	}
}

func TestHealthReporterEmitsHeartbeat(t *testing.T) {
	engine := newEngineBridgeStub()
	ctx := NewPluginRuntimeContext(engine, "rule-test", "node-test", core.HealthOptions{
		HeartbeatInterval: 25 * time.Millisecond,
		MaxSilencePeriod:  time.Minute,
		QueueCapacity:     8,
	}, nil, false)
	if closer, ok := ctx.Health().(interface{ Close() error }); ok {
		t.Cleanup(func() { _ = closer.Close() })
	}

	got := waitForHealthMessages(t, engine, 1)
	if got[0].Kind != core.HealthKindHeartbeat {
		t.Fatalf("expected heartbeat report, got %s", got[0].Kind)
	}
}

func TestRuntimeContextConnectorStreamForwardsToEngineBridge(t *testing.T) {
	engine := newEngineBridgeStub()
	ctx := NewPluginRuntimeContext(engine, "rule-stream", "source-mqtt", core.DefaultHealthOptions(), nil, false)

	stream, err := ctx.Connector().OpenStream(core.StreamOpenRequest{
		ResourceRef:         "mqtt-1",
		Target:              "messages",
		InitialCredits:      8,
		MaxBufferedMessages: 256,
		OverflowPolicy:      "drop_oldest",
		AckTimeoutMS:        90000,
		MaxDeliverBatch:     4,
		Payload:             []byte(`{"topics":["factory/#"]}`),
	})
	if err != nil {
		t.Fatalf("OpenStream failed: %v", err)
	}
	if engine.streamOpenReq.ResourceRef != "mqtt-1" || engine.streamOpenReq.InitialCredits != 8 {
		t.Fatalf("unexpected open request: %#v", engine.streamOpenReq)
	}
	if engine.streamOpenReq.MaxBufferedMessages != 256 || engine.streamOpenReq.OverflowPolicy != "drop_oldest" {
		t.Fatalf("unexpected stream options: %#v", engine.streamOpenReq)
	}
	if engine.streamOpenReq.AckTimeoutMS != 90000 || engine.streamOpenReq.MaxDeliverBatch != 4 {
		t.Fatalf("unexpected stream delivery options: %#v", engine.streamOpenReq)
	}

	message, ok, err := stream.Recv(250 * time.Millisecond)
	if err != nil {
		t.Fatalf("Recv failed: %v", err)
	}
	if !ok {
		t.Fatal("expected a stream message")
	}
	if len(engine.streamRecvReqs) != 1 || engine.streamRecvReqs[0].StreamID != "stream-1" {
		t.Fatalf("unexpected recv requests: %#v", engine.streamRecvReqs)
	}
	if engine.streamRecvReqs[0].MaxMessages != 4 {
		t.Fatalf("expected batched receive size 4, got %#v", engine.streamRecvReqs[0])
	}
	if message.DeliveryID != "delivery-1" {
		t.Fatalf("unexpected stream message: %#v", message)
	}
	second, ok, err := stream.Recv(250 * time.Millisecond)
	if err != nil {
		t.Fatalf("second Recv failed: %v", err)
	}
	if !ok || second.DeliveryID != "delivery-2" {
		t.Fatalf("expected buffered second message, got %#v", second)
	}
	if engine.streamRecvCalls != 1 {
		t.Fatalf("expected buffered receive to use one bridge call, got %d", engine.streamRecvCalls)
	}

	if err := stream.Ack("delivery-1", 2); err != nil {
		t.Fatalf("Ack failed: %v", err)
	}
	if engine.streamAckReq.StreamID != "stream-1" || engine.streamAckReq.Credits != 2 {
		t.Fatalf("unexpected ack request: %#v", engine.streamAckReq)
	}

	if err := stream.Nack("delivery-2", false, "decode_failed", 1); err != nil {
		t.Fatalf("Nack failed: %v", err)
	}
	if engine.streamNackReq.StreamID != "stream-1" || engine.streamNackReq.Credits != 1 {
		t.Fatalf("unexpected nack request: %#v", engine.streamNackReq)
	}

	if err := stream.GrantCredits(4); err != nil {
		t.Fatalf("GrantCredits failed: %v", err)
	}
	if engine.streamGrantReq.StreamID != "stream-1" || engine.streamGrantReq.Credits != 4 {
		t.Fatalf("unexpected grant request: %#v", engine.streamGrantReq)
	}

	if err := stream.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if engine.streamCloseReq.StreamID != "stream-1" {
		t.Fatalf("unexpected close request: %#v", engine.streamCloseReq)
	}
}

func TestNoopConnectorFailsExplicitly(t *testing.T) {
	connector := NewNoopConnectorClient()
	if _, err := connector.Read(core.ConnectorRequest{}); !errors.Is(err, core.ErrNotImplemented) {
		t.Fatalf("Read should fail with ErrNotImplemented, got %v", err)
	}
	if _, err := connector.OpenStream(core.StreamOpenRequest{}); !errors.Is(err, core.ErrNotImplemented) {
		t.Fatalf("OpenStream should fail with ErrNotImplemented, got %v", err)
	}
}

func TestEngineStreamConsumerConcurrentLifecycle(t *testing.T) {
	engine := newEngineBridgeStub()
	ctx := NewPluginRuntimeContext(engine, "rule-stream", "source", core.DefaultHealthOptions(), nil, false)
	stream, err := ctx.Connector().OpenStream(core.StreamOpenRequest{MaxDeliverBatch: 2})
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, _ = stream.Recv(time.Millisecond)
		}()
	}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = stream.Close()
		}()
	}
	wg.Wait()
	if _, _, err := stream.Recv(time.Millisecond); err == nil {
		t.Fatal("Recv after Close should fail")
	}
}

func TestRuntimeContextResourceEventsReturnsNilWhenDisabled(t *testing.T) {
	engine := newEngineBridgeStub()
	ctx := NewPluginRuntimeContext(engine, "rule-test", "node-test", core.DefaultHealthOptions(), nil, false)

	if ctx.ResourceEvents() != nil {
		t.Fatal("expected resource events channel to be nil when disabled")
	}

	time.Sleep(30 * time.Millisecond)
	if got := engine.nextResourceEventCallsCount(); got != 0 {
		t.Fatalf("expected no resource polling when disabled, got %d calls", got)
	}
}

func TestRuntimeContextResourceEventsStartsPollingOnlyAfterRequested(t *testing.T) {
	engine := newEngineBridgeStub()
	ctx := NewPluginRuntimeContext(engine, "rule-test", "node-test", core.DefaultHealthOptions(), nil, true)
	if closer, ok := ctx.(interface{ Close() error }); ok {
		defer func() { _ = closer.Close() }()
	}

	time.Sleep(30 * time.Millisecond)
	if got := engine.nextResourceEventCallsCount(); got != 0 {
		t.Fatalf("expected no polling before ResourceEvents is requested, got %d calls", got)
	}

	if ctx.ResourceEvents() == nil {
		t.Fatal("expected resource events channel when enabled")
	}

	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		if engine.nextResourceEventCallsCount() > 0 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("expected polling to start after ResourceEvents() call")
}

func TestRuntimeContextResourceEventsConcurrentCallsStartSingleLoop(t *testing.T) {
	engine := newEngineBridgeStub()
	block := make(chan struct{})
	engine.nextEvent = func(timeout time.Duration) (core.ResourceStatusEvent, bool, error) {
		<-block
		return core.ResourceStatusEvent{}, false, nil
	}
	ctx := NewPluginRuntimeContext(engine, "rule-test", "node-test", core.DefaultHealthOptions(), nil, true)
	if closer, ok := ctx.(interface{ Close() error }); ok {
		defer func() {
			_ = closer.Close()
			close(block)
		}()
	} else {
		defer close(block)
	}

	results := make([]<-chan core.ResourceStatusEvent, 8)
	var wg sync.WaitGroup
	for i := range results {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index] = ctx.ResourceEvents()
		}(i)
	}
	wg.Wait()

	first := results[0]
	if first == nil {
		t.Fatal("expected non-nil resource events channel")
	}
	for i := 1; i < len(results); i++ {
		if results[i] != first {
			t.Fatalf("expected all callers to receive the same channel, mismatch at index %d", i)
		}
	}

	deadline := time.Now().Add(100 * time.Millisecond)
	for time.Now().Before(deadline) {
		if got := engine.nextResourceEventCallsCount(); got == 1 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("expected a single polling loop, got %d calls", engine.nextResourceEventCallsCount())
}

func TestRuntimeContextResourceEventsBacksOffOnFastEmptyResponses(t *testing.T) {
	engine := newEngineBridgeStub()
	ctx := NewPluginRuntimeContext(engine, "rule-test", "node-test", core.DefaultHealthOptions(), nil, true)
	if closer, ok := ctx.(interface{ Close() error }); ok {
		defer func() { _ = closer.Close() }()
	}

	if ctx.ResourceEvents() == nil {
		t.Fatal("expected resource events channel when enabled")
	}

	time.Sleep(150 * time.Millisecond)
	if got := engine.nextResourceEventCallsCount(); got > 3 {
		t.Fatalf("expected fast empty responses to be throttled, got %d polls", got)
	}
	if !engine.hasCounter("sdk_resource_events_backoff_total") {
		t.Fatal("expected backoff counter to be emitted")
	}
	if !engine.hasObservation("sdk_resource_events_backoff_ms") {
		t.Fatal("expected backoff observation to be emitted")
	}
}
