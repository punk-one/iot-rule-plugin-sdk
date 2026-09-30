package core

import (
	"context"
	"time"
)

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	LogBatch(level LogLevel, messages []string, fields []map[string]interface{})
}

type Emitter interface {
	Publish(event Event) error
	Ack(eventID string) error
	PublishAck(ack AckMessage) error
	EmitTo(label string, event Event) error
	EmitBatch(events []Event) error
}

type Metrics interface {
	IncCounter(name string, labels map[string]string)
	Observe(name string, value float64, labels map[string]string)
}

type StateManager interface {
	GetState(key string, state interface{}) error
	SetState(key string, state interface{}) error
	SetStateWithTTL(key string, state interface{}, ttl time.Duration) error
	DeleteState(key string) error
	GenerateKey(keys []string) string
}

type StateStore interface {
	GetState(ctx context.Context, key string, state interface{}) error
	SetState(ctx context.Context, key string, state interface{}) error
	SetStateWithTTL(ctx context.Context, key string, state interface{}, ttl time.Duration) error
	DeleteState(ctx context.Context, key string) error
}

type StreamConsumer interface {
	ID() string
	Recv(timeout time.Duration) (StreamMessage, bool, error)
	Ack(deliveryID string, credits int) error
	Nack(deliveryID string, requeue bool, reason string, credits int) error
	GrantCredits(credits int) error
	Close() error
}

type ConnectorClient interface {
	Read(req ConnectorRequest) (ConnectorResponse, error)
	Write(req ConnectorRequest) (ConnectorResponse, error)
	BatchRead(req ConnectorRequest) (ConnectorResponse, error)
	BatchWrite(req ConnectorRequest) (ConnectorResponse, error)
	OpenStream(req StreamOpenRequest) (StreamConsumer, error)
	CurrentStatus(resourceRef string) (ResourceStatusEvent, bool)
}

type RuntimeContext interface {
	RuleID() string
	NodeID() string
	Logger() Logger
	Emitter() Emitter
	Metrics() Metrics
	Health() HealthReporter
	Connector() ConnectorClient
	ResourceEvents() <-chan ResourceStatusEvent
}

type Plugin interface {
	Info() PluginInfo
	Init(cfg PluginConfig) error
	Start(ctx RuntimeContext) error
	OnEvent(e Event) error
	OnEvents(events []Event) error
	Stop() error
}

type ConnectorPlugin interface {
	Info() PluginInfo
	CreateResource(resource ConnectorResource) (string, error)
	DestroyResource(providerHandle string) error
	Execute(providerHandle string, req ConnectorRequest) (ConnectorResponse, error)
	OpenStream(providerHandle string, req StreamOpenRequest) (StreamOpenResponse, error)
	ReceiveStream(providerHandle string, req StreamReceiveRequest) (StreamReceiveResponse, error)
	AckStream(providerHandle string, req StreamAckRequest) error
	NackStream(providerHandle string, req StreamNackRequest) error
	GrantCredits(providerHandle string, req StreamGrantCreditsRequest) error
	CloseStream(providerHandle string, req StreamCloseRequest) error
	Probe(providerHandle string, req ConnectorRequest) (ResourceStatusEvent, error)
	Stop() error
}

type BasePlugin struct{}

func (b *BasePlugin) Init(cfg PluginConfig) error    { return nil }
func (b *BasePlugin) Start(ctx RuntimeContext) error { return nil }
func (b *BasePlugin) Stop() error                    { return nil }
func (b *BasePlugin) OnEvent(e Event) error          { return ErrNotImplemented }
func (b *BasePlugin) OnEvents(events []Event) error  { return ErrBatchNotSupported }
