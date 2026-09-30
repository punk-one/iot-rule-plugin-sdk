package sdk

import (
	"context"
	"fmt"
	"log"

	"github.com/punk-one/iot-rule-plugin-sdk/internal/core"
	internalruntime "github.com/punk-one/iot-rule-plugin-sdk/internal/runtime"
)

// RuntimeContext 是插件运行期上下文的稳定入口。
type RuntimeContext = core.RuntimeContext
type Logger = core.Logger
type Emitter = core.Emitter
type Metrics = core.Metrics

const (
	AckStatusSuccess = core.AckStatusSuccess
	AckStatusFailed  = core.AckStatusFailed
)

type AckMessage = core.AckMessage

func NewRPCLoggerFromEngine(engine EngineRPC, ruleID, nodeID string) Logger {
	return internalruntime.NewRPCLoggerFromEngine(engine, ruleID, nodeID)
}

// DefaultLogger 默认日志实现
// 使用标准库 log 包，避免依赖外部包
type DefaultLogger struct{}

func (l *DefaultLogger) Debug(msg string, fields ...Field) {
	log.Printf("[DEBUG] %s %v", msg, fields)
}

func (l *DefaultLogger) Info(msg string, fields ...Field) {
	log.Printf("[INFO] %s %v", msg, fields)
}

func (l *DefaultLogger) Warn(msg string, fields ...Field) {
	log.Printf("[WARN] %s %v", msg, fields)
}

func (l *DefaultLogger) Error(msg string, fields ...Field) {
	log.Printf("[ERROR] %s %v", msg, fields)
}

func (l *DefaultLogger) LogBatch(level LogLevel, messages []string, fields []map[string]interface{}) {
	for i, msg := range messages {
		var fieldList []Field
		if i < len(fields) {
			for key, value := range fields[i] {
				fieldList = append(fieldList, Field{Key: key, Value: value})
			}
		}
		switch level {
		case LogLevelDebug:
			l.Debug(msg, fieldList...)
		case LogLevelInfo:
			l.Info(msg, fieldList...)
		case LogLevelWarn:
			l.Warn(msg, fieldList...)
		case LogLevelError:
			l.Error(msg, fieldList...)
		}
	}
}

// EmitterAdapter 将 Publish 函数适配为 Emitter 接口
type EmitterAdapter struct {
	publisher func(port int, event Event) error
}

func (e *EmitterAdapter) Publish(event Event) error {
	if e.publisher == nil {
		return fmt.Errorf("publisher not configured")
	}
	return e.publisher(0, event)
}

func (e *EmitterAdapter) Ack(eventID string) error { return nil }

func (e *EmitterAdapter) PublishAck(ack AckMessage) error { return nil }

func (e *EmitterAdapter) EmitTo(label string, event Event) error {
	if e.publisher == nil {
		return fmt.Errorf("publisher not configured")
	}
	if event.Metadata == nil {
		event.Metadata = make(map[string]string)
	}
	event.Metadata["route_label"] = label
	return e.publisher(0, event)
}

func (e *EmitterAdapter) EmitBatch(events []Event) error {
	if e.publisher == nil {
		return fmt.Errorf("publisher not configured")
	}
	for _, event := range events {
		if err := e.publisher(0, event); err != nil {
			return err
		}
	}
	return nil
}

func NewRPCEmitterFromEngine(engine EngineRPC) Emitter {
	return internalruntime.NewRPCEmitterFromEngine(engine)
}

// DefaultMetrics 默认指标实现（空实现）
type DefaultMetrics = internalruntime.DefaultMetrics

func NewRPCMetricsFromEngine(engine EngineRPC, ruleID, nodeID string) Metrics {
	return internalruntime.NewRPCMetricsFromEngine(engine, ruleID, nodeID)
}

// BuiltinRuntimeContext 是内置插件可选依赖的扩展上下文。
// 外部插件通常拿不到此能力。
type BuiltinRuntimeContext interface {
	RuntimeContext
	Context() context.Context
}

// StatefulRuntimeContext 是需要状态管理能力的插件可选依赖上下文。
type StatefulRuntimeContext interface {
	RuntimeContext
	State() StateManager
}

// EngineBridgeRuntimeContext 是引擎与插件进程桥接时使用的扩展上下文。
// 仅 Engine 内部的加载链路应依赖该接口。
type EngineBridgeRuntimeContext interface {
	RuntimeContext
	GetEngineRPCServer() interface{}
	SavePluginRPCClient(client *PluginRPCClient)
}

// RuntimeContextImpl RuntimeContext 实现（用于内置插件）
type RuntimeContextImpl struct {
	ruleID          string
	nodeID          string
	ctx             context.Context
	logger          Logger
	metrics         Metrics
	emitter         Emitter
	health          HealthReporter
	stateMgr        StateManager
	connector       ConnectorClient
	resourceEvents  <-chan ResourceStatusEvent
	savePluginRPCCB func(client *PluginRPCClient)
}

func (r *RuntimeContextImpl) RuleID() string   { return r.ruleID }
func (r *RuntimeContextImpl) NodeID() string   { return r.nodeID }
func (r *RuntimeContextImpl) Logger() Logger   { return r.logger }
func (r *RuntimeContextImpl) Emitter() Emitter { return r.emitter }
func (r *RuntimeContextImpl) Metrics() Metrics { return r.metrics }

func (r *RuntimeContextImpl) Health() HealthReporter {
	if r.health == nil {
		r.health = newNoopHealthReporter()
	}
	return r.health
}

func (r *RuntimeContextImpl) Connector() ConnectorClient {
	if r.connector == nil {
		r.connector = internalruntime.NewNoopConnectorClient()
	}
	return r.connector
}

func (r *RuntimeContextImpl) ResourceEvents() <-chan ResourceStatusEvent { return r.resourceEvents }
func (r *RuntimeContextImpl) Context() context.Context                   { return r.ctx }
func (r *RuntimeContextImpl) State() StateManager                        { return r.stateMgr }

func (r *RuntimeContextImpl) SavePluginRPCClient(client *PluginRPCClient) {
	if r.savePluginRPCCB != nil {
		r.savePluginRPCCB(client)
	}
}
