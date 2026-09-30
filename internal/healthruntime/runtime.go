package healthruntime

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Status string

const (
	StatusUnknown      Status = "unknown"
	StatusInitializing Status = "initializing"
	StatusHealthy      Status = "healthy"
	StatusDegraded     Status = "degraded"
	StatusRecovering   Status = "recovering"
	StatusUnhealthy    Status = "unhealthy"
	StatusStopped      Status = "stopped"
)

type Kind string

const (
	KindState     Kind = "state"
	KindHeartbeat Kind = "heartbeat"
)

const (
	defaultCloseTimeout = 10 * time.Second
	sendRetryFloor      = 100 * time.Millisecond
	sendRetryMax        = time.Second
)

var ErrReporterClosed = errors.New("health reporter is closed")

type Options struct {
	HeartbeatInterval time.Duration `json:"heartbeat_interval,omitempty"`
	MaxSilencePeriod  time.Duration `json:"max_silence_period,omitempty"`
	QueueCapacity     int           `json:"queue_capacity,omitempty"`
}

func DefaultOptions() Options {
	return Options{HeartbeatInterval: 15 * time.Second, MaxSilencePeriod: time.Minute, QueueCapacity: 32}
}

func (o Options) Normalize() Options {
	defaults := DefaultOptions()
	if o.HeartbeatInterval <= 0 {
		o.HeartbeatInterval = defaults.HeartbeatInterval
	}
	if o.MaxSilencePeriod <= 0 {
		o.MaxSilencePeriod = defaults.MaxSilencePeriod
	}
	if o.QueueCapacity <= 0 {
		o.QueueCapacity = defaults.QueueCapacity
	}
	return o
}

type Report struct {
	Status     Status            `json:"status"`
	Code       string            `json:"code,omitempty"`
	Message    string            `json:"message,omitempty"`
	Details    map[string]string `json:"details,omitempty"`
	ObservedAt time.Time         `json:"observed_at,omitempty"`
	Sequence   uint64            `json:"sequence,omitempty"`
}

type Heartbeat struct {
	Message    string            `json:"message,omitempty"`
	Details    map[string]string `json:"details,omitempty"`
	ObservedAt time.Time         `json:"observed_at,omitempty"`
	Sequence   uint64            `json:"sequence,omitempty"`
}

type ReportArgs struct {
	Kind               Kind              `json:"kind"`
	Status             Status            `json:"status,omitempty"`
	Code               string            `json:"code,omitempty"`
	Message            string            `json:"message,omitempty"`
	Details            map[string]string `json:"details,omitempty"`
	ObservedAtUnixNano int64             `json:"observed_at_unix_nano,omitempty"`
	Sequence           uint64            `json:"sequence,omitempty"`
}

type Reporter interface {
	Report(report Report) error
	Healthy(msg string, details map[string]string) error
	Degraded(code, msg string, details map[string]string) error
	Recovering(code, msg string, details map[string]string) error
	Unhealthy(code, msg string, details map[string]string) error
	Heartbeat(msg string, details map[string]string) error
}

type Sender interface {
	SendHealth(args ReportArgs) error
}

// Stats exposes monotonic diagnostics without changing the stable Reporter interface.
type Stats struct {
	CoalescedStates     uint64 `json:"coalesced_states"`
	CoalescedHeartbeats uint64 `json:"coalesced_heartbeats"`
	SenderErrors        uint64 `json:"sender_errors"`
}

type healthEnvelope struct {
	args     ReportArgs
	stateKey string
}

type asyncReporter struct {
	sender  Sender
	options Options

	queue       chan healthEnvelope
	heartbeatCh chan healthEnvelope
	stopCh      chan struct{}
	abortCh     chan struct{}
	doneCh      chan struct{}

	sequence    atomic.Uint64
	outstanding atomic.Int64

	coalescedStates     atomic.Uint64
	coalescedHeartbeats atomic.Uint64
	senderErrors        atomic.Uint64

	mu                    sync.Mutex
	lastSentStateKey      string
	lastSentStateAt       time.Time
	lastAcceptedStateKey  string
	lastAcceptedStateAt   time.Time
	lastSentHeartbeat     time.Time
	lastAcceptedHeartbeat time.Time
	pendingState          *healthEnvelope
	closed                bool
	closeOnce             sync.Once
	abortOnce             sync.Once
}

type nopReporter struct{}

func (nopReporter) Report(Report) error                                { return nil }
func (nopReporter) Healthy(string, map[string]string) error            { return nil }
func (nopReporter) Degraded(string, string, map[string]string) error   { return nil }
func (nopReporter) Recovering(string, string, map[string]string) error { return nil }
func (nopReporter) Unhealthy(string, string, map[string]string) error  { return nil }
func (nopReporter) Heartbeat(string, map[string]string) error          { return nil }

func NewNoop() Reporter { return nopReporter{} }

func NewAsync(sender Sender, options Options) Reporter {
	if sender == nil {
		return NewNoop()
	}
	reporter := &asyncReporter{
		sender:      sender,
		options:     options.Normalize(),
		heartbeatCh: make(chan healthEnvelope, 1),
		stopCh:      make(chan struct{}),
		abortCh:     make(chan struct{}),
		doneCh:      make(chan struct{}),
	}
	reporter.queue = make(chan healthEnvelope, reporter.options.QueueCapacity)
	go reporter.run()
	return reporter
}

func (r *asyncReporter) Close() error {
	if r == nil {
		return nil
	}
	r.closeOnce.Do(func() {
		r.mu.Lock()
		r.closed = true
		close(r.stopCh)
		r.mu.Unlock()
	})
	timer := time.NewTimer(defaultCloseTimeout)
	defer timer.Stop()
	select {
	case <-r.doneCh:
		return nil
	case <-timer.C:
		r.abortOnce.Do(func() { close(r.abortCh) })
		return fmt.Errorf("flush health reports: timeout after %s", defaultCloseTimeout)
	}
}

func (r *asyncReporter) Flush(timeout time.Duration) error {
	if r == nil || timeout <= 0 {
		return nil
	}
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		if r.isDrained() {
			return nil
		}
		select {
		case <-r.doneCh:
			if r.isDrained() {
				return nil
			}
			return ErrReporterClosed
		case <-ticker.C:
		case <-timer.C:
			return fmt.Errorf("flush health reports: timeout after %s", timeout)
		}
	}
}

func (r *asyncReporter) Stats() Stats {
	if r == nil {
		return Stats{}
	}
	return Stats{
		CoalescedStates:     r.coalescedStates.Load(),
		CoalescedHeartbeats: r.coalescedHeartbeats.Load(),
		SenderErrors:        r.senderErrors.Load(),
	}
}

func Flush(reporter Reporter, timeout time.Duration) error {
	if flusher, ok := reporter.(interface{ Flush(time.Duration) error }); ok {
		return flusher.Flush(timeout)
	}
	return nil
}

func Snapshot(reporter Reporter) Stats {
	if provider, ok := reporter.(interface{ Stats() Stats }); ok {
		return provider.Stats()
	}
	return Stats{}
}

func (r *asyncReporter) Report(report Report) error {
	if report.Status == "" {
		report.Status = StatusUnknown
	}
	if report.ObservedAt.IsZero() {
		report.ObservedAt = time.Now()
	}
	report.Details = cloneDetails(report.Details)
	report.Sequence = r.sequence.Add(1)
	stateKey := buildStateKey(report)
	envelope := healthEnvelope{stateKey: stateKey, args: ReportArgs{
		Kind: KindState, Status: report.Status, Code: report.Code, Message: report.Message,
		Details: report.Details, ObservedAtUnixNano: report.ObservedAt.UnixNano(), Sequence: report.Sequence,
	}}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ErrReporterClosed
	}
	if r.shouldSuppressState(stateKey, report.ObservedAt) {
		return nil
	}
	r.lastAcceptedStateKey = stateKey
	r.lastAcceptedStateAt = report.ObservedAt
	r.lastAcceptedHeartbeat = report.ObservedAt
	if r.pendingState != nil {
		r.pendingState = &envelope
		r.coalescedStates.Add(1)
		return nil
	}
	r.outstanding.Add(1)
	select {
	case r.queue <- envelope:
	default:
		r.pendingState = &envelope
		r.coalescedStates.Add(1)
	}
	return nil
}

func (r *asyncReporter) shouldSuppressState(stateKey string, observedAt time.Time) bool {
	if r.lastAcceptedStateKey == stateKey && !r.lastAcceptedStateAt.IsZero() && observedAt.Sub(r.lastAcceptedStateAt) < r.options.MaxSilencePeriod {
		return true
	}
	return r.lastSentStateKey == stateKey && !r.lastSentStateAt.IsZero() && observedAt.Sub(r.lastSentStateAt) < r.options.MaxSilencePeriod
}

func (r *asyncReporter) Healthy(msg string, details map[string]string) error {
	return r.Report(Report{Status: StatusHealthy, Message: msg, Details: details})
}
func (r *asyncReporter) Degraded(code, msg string, details map[string]string) error {
	return r.Report(Report{Status: StatusDegraded, Code: code, Message: msg, Details: details})
}
func (r *asyncReporter) Recovering(code, msg string, details map[string]string) error {
	return r.Report(Report{Status: StatusRecovering, Code: code, Message: msg, Details: details})
}
func (r *asyncReporter) Unhealthy(code, msg string, details map[string]string) error {
	return r.Report(Report{Status: StatusUnhealthy, Code: code, Message: msg, Details: details})
}

func (r *asyncReporter) Heartbeat(msg string, details map[string]string) error {
	observedAt := time.Now()
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return ErrReporterClosed
	}
	if (!r.lastAcceptedHeartbeat.IsZero() && observedAt.Sub(r.lastAcceptedHeartbeat) < r.options.HeartbeatInterval) ||
		(!r.lastSentHeartbeat.IsZero() && observedAt.Sub(r.lastSentHeartbeat) < r.options.HeartbeatInterval) {
		r.mu.Unlock()
		return nil
	}
	r.lastAcceptedHeartbeat = observedAt
	r.mu.Unlock()

	envelope := healthEnvelope{args: ReportArgs{
		Kind: KindHeartbeat, Message: msg, Details: cloneDetails(details),
		ObservedAtUnixNano: observedAt.UnixNano(), Sequence: r.sequence.Add(1),
	}}
	r.outstanding.Add(1)
	select {
	case r.heartbeatCh <- envelope:
	default:
		select {
		case <-r.heartbeatCh:
			r.outstanding.Add(-1)
		default:
		}
		r.heartbeatCh <- envelope
		r.coalescedHeartbeats.Add(1)
	}
	return nil
}

func (r *asyncReporter) run() {
	defer close(r.doneCh)
	ticker := time.NewTicker(r.options.HeartbeatInterval)
	defer ticker.Stop()
	stopCh := r.stopCh
	closing := false
	for {
		if item, ok := r.nextImmediate(); ok {
			if !r.sendUntilSuccess(item) {
				return
			}
			continue
		}
		if closing {
			return
		}
		select {
		case item := <-r.queue:
			if !r.sendUntilSuccess(item) {
				return
			}
		case item := <-r.heartbeatCh:
			if !r.sendUntilSuccess(item) {
				return
			}
		case <-ticker.C:
			_ = r.Heartbeat("", nil)
		case <-stopCh:
			closing = true
			stopCh = nil
		}
	}
}

func (r *asyncReporter) nextImmediate() (healthEnvelope, bool) {
	select {
	case item := <-r.queue:
		return item, true
	default:
	}
	r.mu.Lock()
	if r.pendingState != nil {
		item := *r.pendingState
		r.pendingState = nil
		r.mu.Unlock()
		return item, true
	}
	r.mu.Unlock()
	select {
	case item := <-r.heartbeatCh:
		return item, true
	default:
		return healthEnvelope{}, false
	}
}

func (r *asyncReporter) sendUntilSuccess(item healthEnvelope) bool {
	backoff := sendRetryFloor
	for {
		err := r.sender.SendHealth(item.args)
		if err == nil {
			r.markSent(item)
			r.outstanding.Add(-1)
			return true
		}
		r.senderErrors.Add(1)
		timer := time.NewTimer(backoff)
		select {
		case <-r.abortCh:
			timer.Stop()
			return false
		case <-timer.C:
		}
		if backoff < sendRetryMax {
			backoff *= 2
			if backoff > sendRetryMax {
				backoff = sendRetryMax
			}
		}
	}
}

func (r *asyncReporter) markSent(item healthEnvelope) {
	observedAt := time.Unix(0, item.args.ObservedAtUnixNano)
	r.mu.Lock()
	defer r.mu.Unlock()
	if item.args.Kind == KindState {
		r.lastSentStateKey = item.stateKey
		r.lastSentStateAt = observedAt
	}
	r.lastSentHeartbeat = observedAt
}

func (r *asyncReporter) isDrained() bool {
	return r.outstanding.Load() == 0
}

func buildStateKey(report Report) string {
	detailsJSON, _ := json.Marshal(report.Details)
	return string(report.Status) + "|" + report.Code + "|" + report.Message + "|" + string(detailsJSON)
}

func cloneDetails(details map[string]string) map[string]string {
	if len(details) == 0 {
		return nil
	}
	out := make(map[string]string, len(details))
	for key, value := range details {
		out[key] = value
	}
	return out
}
