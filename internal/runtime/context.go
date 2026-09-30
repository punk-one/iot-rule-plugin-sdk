package runtime

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/punk-one/iot-rule-plugin-sdk/internal/core"
)

type EngineBridge interface {
	Emit(e core.Event) error
	EmitWithTargets(e core.Event, toNodeIDs []string) error
	Log(level core.LogLevel, msg string, fields map[string]interface{})
	LogBatch(level core.LogLevel, messages []string, fields []map[string]interface{})
	IncCounter(name string, labels map[string]string)
	Observe(name string, value float64, labels map[string]string)
	ReportHealth(args core.ReportHealthArgs) error
	Ack(eventID string) error
	PublishAck(ack core.AckMessage) error
	EmitBatch(events []core.Event) error
	GetState(key string) ([]byte, error)
	SetState(key string, value []byte) error
	SetStateWithTTL(key string, value []byte, ttl time.Duration) error
	DeleteState(key string) error
	ExecuteConnector(req core.ConnectorRequest) (core.ConnectorResponse, error)
	OpenConnectorStream(req core.StreamOpenRequest) (core.StreamOpenResponse, error)
	ReceiveConnectorStream(req core.StreamReceiveRequest) (core.StreamReceiveResponse, error)
	AckConnectorStream(req core.StreamAckRequest) error
	NackConnectorStream(req core.StreamNackRequest) error
	GrantConnectorStream(req core.StreamGrantCreditsRequest) error
	CloseConnectorStream(req core.StreamCloseRequest) error
	CurrentResourceStatus(resourceRef string) (core.ResourceStatusEvent, bool)
	NextResourceEvent(timeout time.Duration) (core.ResourceStatusEvent, bool, error)
}

type PluginRuntimeContext struct {
	ruleID                string
	nodeID                string
	engine                EngineBridge
	logger                core.Logger
	emitter               core.Emitter
	metrics               core.Metrics
	health                core.HealthReporter
	stateMgr              core.StateManager
	connector             core.ConnectorClient
	resourceEventsEnabled bool
	resourceEvents        chan core.ResourceStatusEvent
	resourceEventsInit    sync.Once
	resourceEventsMu      sync.RWMutex
	stopCh                chan struct{}
	stopOnce              sync.Once
	closed                atomic.Bool
	loopStarted           atomic.Bool
}

type noopConnectorClient struct{}
type engineConnectorClient struct {
	engine EngineBridge
	ruleID string
	nodeID string
}

type engineStreamConsumer struct {
	engine          EngineBridge
	streamID        string
	maxDeliverBatch int
	recvMu          sync.Mutex
	buffer          []core.StreamMessage
	closed          atomic.Bool
	closeOnce       sync.Once
	closeErr        error
}

const (
	resourceEventPollTimeout       = time.Second
	resourceEventRetryFloor        = 100 * time.Millisecond
	resourceEventRetryMax          = time.Second
	resourceEventWarnThreshold     = 500 * time.Millisecond
	resourceEventTimeoutGrace      = 100 * time.Millisecond
	resourceEventBackoffCounter    = "sdk_resource_events_backoff_total"
	resourceEventBackoffObserve    = "sdk_resource_events_backoff_ms"
	resourceEventFastReturnCounter = "sdk_resource_events_fast_return_total"
)

func NewNoopConnectorClient() core.ConnectorClient {
	return noopConnectorClient{}
}

func (noopConnectorClient) Read(req core.ConnectorRequest) (core.ConnectorResponse, error) {
	return core.ConnectorResponse{}, core.ErrNotImplemented
}

func (noopConnectorClient) Write(req core.ConnectorRequest) (core.ConnectorResponse, error) {
	return core.ConnectorResponse{}, core.ErrNotImplemented
}

func (noopConnectorClient) BatchRead(req core.ConnectorRequest) (core.ConnectorResponse, error) {
	return core.ConnectorResponse{}, core.ErrNotImplemented
}

func (noopConnectorClient) BatchWrite(req core.ConnectorRequest) (core.ConnectorResponse, error) {
	return core.ConnectorResponse{}, core.ErrNotImplemented
}

func (noopConnectorClient) OpenStream(req core.StreamOpenRequest) (core.StreamConsumer, error) {
	return nil, core.ErrNotImplemented
}

func (noopConnectorClient) CurrentStatus(resourceRef string) (core.ResourceStatusEvent, bool) {
	return core.ResourceStatusEvent{}, false
}

func (c engineConnectorClient) Read(req core.ConnectorRequest) (core.ConnectorResponse, error) {
	req.Kind = core.ConnectorRequestRead
	return c.execute(req)
}

func (c engineConnectorClient) Write(req core.ConnectorRequest) (core.ConnectorResponse, error) {
	req.Kind = core.ConnectorRequestWrite
	return c.execute(req)
}

func (c engineConnectorClient) BatchRead(req core.ConnectorRequest) (core.ConnectorResponse, error) {
	req.Kind = core.ConnectorRequestBatchRead
	return c.execute(req)
}

func (c engineConnectorClient) BatchWrite(req core.ConnectorRequest) (core.ConnectorResponse, error) {
	req.Kind = core.ConnectorRequestBatchWrite
	return c.execute(req)
}

func (c engineConnectorClient) OpenStream(req core.StreamOpenRequest) (core.StreamConsumer, error) {
	if c.engine == nil {
		return nil, errors.New("engine connector bridge is not available")
	}
	if req.SessionKey == "" {
		req.SessionKey = c.defaultSessionKey(req.ResourceRef)
	}
	if req.InitialCredits <= 0 {
		req.InitialCredits = 1
	}
	resp, err := c.engine.OpenConnectorStream(req)
	if err != nil {
		return nil, err
	}
	batchSize := req.MaxDeliverBatch
	if batchSize <= 0 {
		batchSize = 1
	}
	return &engineStreamConsumer{
		engine:          c.engine,
		streamID:        resp.StreamID,
		maxDeliverBatch: batchSize,
	}, nil
}

func (c engineConnectorClient) CurrentStatus(resourceRef string) (core.ResourceStatusEvent, bool) {
	if c.engine == nil {
		return core.ResourceStatusEvent{}, false
	}
	return c.engine.CurrentResourceStatus(resourceRef)
}

func (c engineConnectorClient) execute(req core.ConnectorRequest) (core.ConnectorResponse, error) {
	if c.engine == nil {
		return core.ConnectorResponse{}, errors.New("engine connector bridge is not available")
	}
	return c.engine.ExecuteConnector(req)
}

func (c engineConnectorClient) defaultSessionKey(resourceRef string) string {
	return c.ruleID + ":" + c.nodeID + ":" + resourceRef
}

func (c *engineStreamConsumer) ID() string {
	if c == nil {
		return ""
	}
	return c.streamID
}

func (c *engineStreamConsumer) Recv(timeout time.Duration) (core.StreamMessage, bool, error) {
	if c == nil || c.engine == nil {
		return core.StreamMessage{}, false, errors.New("engine stream bridge is not available")
	}
	if c.closed.Load() {
		return core.StreamMessage{}, false, errors.New("engine stream is closed")
	}
	c.recvMu.Lock()
	defer c.recvMu.Unlock()
	if c.closed.Load() {
		return core.StreamMessage{}, false, errors.New("engine stream is closed")
	}
	if len(c.buffer) > 0 {
		message := c.buffer[0]
		c.buffer = c.buffer[1:]
		return message, true, nil
	}
	maxMessages := c.maxDeliverBatch
	if maxMessages <= 0 {
		maxMessages = 1
	}
	resp, err := c.engine.ReceiveConnectorStream(core.StreamReceiveRequest{
		StreamID:      c.streamID,
		MaxMessages:   maxMessages,
		WaitTimeoutMS: int(timeout / time.Millisecond),
	})
	if err != nil {
		return core.StreamMessage{}, false, err
	}
	if len(resp.Messages) == 0 {
		return core.StreamMessage{}, false, nil
	}
	if len(resp.Messages) > 1 {
		c.buffer = append(c.buffer[:0], resp.Messages[1:]...)
	}
	return resp.Messages[0], true, nil
}

func (c *engineStreamConsumer) Ack(deliveryID string, credits int) error {
	if c == nil || c.engine == nil {
		return errors.New("engine stream bridge is not available")
	}
	if c.closed.Load() {
		return errors.New("engine stream is closed")
	}
	return c.engine.AckConnectorStream(core.StreamAckRequest{
		StreamID:   c.streamID,
		DeliveryID: deliveryID,
		Credits:    credits,
	})
}

func (c *engineStreamConsumer) Nack(deliveryID string, requeue bool, reason string, credits int) error {
	if c == nil || c.engine == nil {
		return errors.New("engine stream bridge is not available")
	}
	if c.closed.Load() {
		return errors.New("engine stream is closed")
	}
	return c.engine.NackConnectorStream(core.StreamNackRequest{
		StreamID:   c.streamID,
		DeliveryID: deliveryID,
		Reason:     reason,
		Requeue:    requeue,
		Credits:    credits,
	})
}

func (c *engineStreamConsumer) GrantCredits(credits int) error {
	if c == nil || c.engine == nil {
		return errors.New("engine stream bridge is not available")
	}
	if c.closed.Load() {
		return errors.New("engine stream is closed")
	}
	return c.engine.GrantConnectorStream(core.StreamGrantCreditsRequest{
		StreamID: c.streamID,
		Credits:  credits,
	})
}

func (c *engineStreamConsumer) Close() error {
	if c == nil || c.engine == nil {
		return nil
	}
	c.closeOnce.Do(func() {
		c.closed.Store(true)
		c.closeErr = c.engine.CloseConnectorStream(core.StreamCloseRequest{StreamID: c.streamID})
	})
	return c.closeErr
}

func NewPluginRuntimeContext(engine EngineBridge, ruleID, nodeID string, healthOptions core.HealthOptions, stateMgr core.StateManager, resourceEventsEnabled bool) core.RuntimeContext {
	if stateMgr == nil {
		stateMgr = NewStateManager(ruleID, nodeID, NewRPCStateStore(engine))
	}
	ctx := &PluginRuntimeContext{
		ruleID:                ruleID,
		nodeID:                nodeID,
		engine:                engine,
		logger:                NewRPCLoggerFromEngine(engine, ruleID, nodeID),
		emitter:               NewRPCEmitterFromEngine(engine),
		metrics:               NewRPCMetricsFromEngine(engine, ruleID, nodeID),
		health:                NewEngineHealthReporter(engine, healthOptions),
		stateMgr:              stateMgr,
		connector:             engineConnectorClient{engine: engine, ruleID: ruleID, nodeID: nodeID},
		resourceEventsEnabled: resourceEventsEnabled && engine != nil,
		stopCh:                make(chan struct{}),
	}
	if engine == nil {
		ctx.connector = NewNoopConnectorClient()
	}
	return ctx
}

func (r *PluginRuntimeContext) RuleID() string                  { return r.ruleID }
func (r *PluginRuntimeContext) NodeID() string                  { return r.nodeID }
func (r *PluginRuntimeContext) Logger() core.Logger             { return r.logger }
func (r *PluginRuntimeContext) Emitter() core.Emitter           { return r.emitter }
func (r *PluginRuntimeContext) Metrics() core.Metrics           { return r.metrics }
func (r *PluginRuntimeContext) Health() core.HealthReporter     { return r.health }
func (r *PluginRuntimeContext) State() core.StateManager        { return r.stateMgr }
func (r *PluginRuntimeContext) Connector() core.ConnectorClient { return r.connector }
func (r *PluginRuntimeContext) ResourceEvents() <-chan core.ResourceStatusEvent {
	if r == nil || !r.resourceEventsEnabled || r.closed.Load() {
		return nil
	}

	r.resourceEventsInit.Do(func() {
		if r.closed.Load() {
			return
		}
		ch := make(chan core.ResourceStatusEvent, 8)
		r.resourceEventsMu.Lock()
		defer r.resourceEventsMu.Unlock()
		if r.closed.Load() {
			return
		}
		r.resourceEvents = ch
		r.loopStarted.Store(true)
		go r.resourceEventLoop(ch)
	})

	if r.closed.Load() {
		return nil
	}
	r.resourceEventsMu.RLock()
	defer r.resourceEventsMu.RUnlock()
	return r.resourceEvents
}

func (r *PluginRuntimeContext) Close() error {
	if r == nil || r.stopCh == nil {
		return nil
	}
	r.closed.Store(true)
	r.stopOnce.Do(func() {
		close(r.stopCh)
		if closer, ok := r.logger.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	})
	return nil
}

func (r *PluginRuntimeContext) resourceEventLoop(events chan core.ResourceStatusEvent) {
	if r == nil || r.engine == nil {
		return
	}
	defer close(events)

	var backoff time.Duration
	var backoffActive bool
	var warned bool
	consecutiveFailures := 0

	for {
		select {
		case <-r.stopCh:
			return
		default:
		}

		start := time.Now()
		event, ok, err := r.engine.NextResourceEvent(resourceEventPollTimeout)
		elapsed := time.Since(start)

		if err != nil || !ok {
			delay, reason := nextResourceEventRetryDelay(backoff, elapsed, err, ok)
			if delay > 0 {
				consecutiveFailures++
				labels := resourceEventMetricLabels(r.ruleID, r.nodeID, reason)
				if !backoffActive {
					r.metrics.IncCounter(resourceEventBackoffCounter, labels)
					if reason == "fast_false" {
						r.metrics.IncCounter(resourceEventFastReturnCounter, labels)
					}
				}
				if delay != backoff {
					r.metrics.Observe(resourceEventBackoffObserve, float64(delay/time.Millisecond), labels)
				}
				if delay >= resourceEventWarnThreshold && !warned {
					r.logger.Warn("Resource events loop entering backoff",
						core.Field{Key: "rule_id", Value: r.ruleID},
						core.Field{Key: "node_id", Value: r.nodeID},
						core.Field{Key: "reason", Value: reason},
						core.Field{Key: "backoff_ms", Value: delay.Milliseconds()},
						core.Field{Key: "consecutive_failures", Value: consecutiveFailures},
					)
					warned = true
				}
				backoff = delay
				backoffActive = true
				if !waitForRetryOrStop(r.stopCh, delay) {
					return
				}
			} else if backoffActive {
				r.logger.Info("Resource events loop recovered",
					core.Field{Key: "rule_id", Value: r.ruleID},
					core.Field{Key: "node_id", Value: r.nodeID},
				)
				backoff = 0
				backoffActive = false
				warned = false
				consecutiveFailures = 0
			}
			continue
		}

		if backoffActive {
			r.logger.Info("Resource events loop recovered",
				core.Field{Key: "rule_id", Value: r.ruleID},
				core.Field{Key: "node_id", Value: r.nodeID},
			)
			backoff = 0
			backoffActive = false
			warned = false
			consecutiveFailures = 0
		}

		select {
		case <-r.stopCh:
			return
		case events <- event:
		default:
			select {
			case <-events:
			default:
			}
			events <- event
		}
	}
}

func nextResourceEventRetryDelay(previous, elapsed time.Duration, err error, ok bool) (time.Duration, string) {
	if err == nil && !ok && elapsed+resourceEventTimeoutGrace >= resourceEventPollTimeout {
		return 0, ""
	}

	reason := "fast_false"
	if err != nil {
		reason = "rpc_error"
	}
	if previous < resourceEventRetryFloor {
		return resourceEventRetryFloor, reason
	}
	next := previous * 2
	if next > resourceEventRetryMax {
		next = resourceEventRetryMax
	}
	return next, reason
}

func waitForRetryOrStop(stopCh <-chan struct{}, delay time.Duration) bool {
	if delay <= 0 {
		return true
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-stopCh:
		return false
	case <-timer.C:
		return true
	}
}

func resourceEventMetricLabels(ruleID, nodeID, reason string) map[string]string {
	return map[string]string{
		"rule_id": ruleID,
		"node_id": nodeID,
		"reason":  reason,
	}
}
