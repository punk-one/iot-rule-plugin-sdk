package runtime

import (
	"reflect"
	"sync"
	"time"

	"github.com/punk-one/iot-rule-plugin-sdk/internal/core"
)

type EngineRPCLogger struct {
	engine    EngineBridge
	ruleID    string
	nodeID    string
	mu        sync.Mutex
	logCache  map[core.LogLevel]*logBuffer
	stopCh    chan struct{}
	doneCh    chan struct{}
	closeOnce sync.Once
	closed    bool
}

type logBuffer struct {
	messages []string
	fields   []map[string]interface{}
}

func NewRPCLoggerFromEngine(engine EngineBridge, ruleID, nodeID string) core.Logger {
	logger := &EngineRPCLogger{
		engine:   engine,
		ruleID:   ruleID,
		nodeID:   nodeID,
		logCache: make(map[core.LogLevel]*logBuffer),
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
	levels := []core.LogLevel{core.LogLevelDebug, core.LogLevelInfo, core.LogLevelWarn, core.LogLevelError}
	for _, level := range levels {
		logger.logCache[level] = &logBuffer{
			messages: make([]string, 0, 100),
			fields:   make([]map[string]interface{}, 0, 100),
		}
	}
	go logger.startFlusher()
	return logger
}

func (l *EngineRPCLogger) startFlusher() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	defer close(l.doneCh)
	for {
		select {
		case <-ticker.C:
			l.Flush()
		case <-l.stopCh:
			l.Flush()
			return
		}
	}
}

func (l *EngineRPCLogger) Flush() {
	type pendingBatch struct {
		level    core.LogLevel
		messages []string
		fields   []map[string]interface{}
	}

	l.mu.Lock()
	batches := make([]pendingBatch, 0, len(l.logCache))
	for level, buf := range l.logCache {
		if len(buf.messages) == 0 {
			continue
		}
		batches = append(batches, pendingBatch{
			level:    level,
			messages: buf.messages,
			fields:   buf.fields,
		})
		buf.messages = make([]string, 0, 100)
		buf.fields = make([]map[string]interface{}, 0, 100)
	}
	l.mu.Unlock()

	if l.engine == nil {
		return
	}
	for _, batch := range batches {
		l.engine.LogBatch(batch.level, batch.messages, batch.fields)
	}
}

func (l *EngineRPCLogger) addLog(level core.LogLevel, msg string, fields ...core.Field) {
	fieldsMap := make(map[string]interface{})
	for _, field := range fields {
		fieldsMap[field.Key] = cloneLogValue(field.Value)
	}
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return
	}
	buf := l.logCache[level]
	buf.messages = append(buf.messages, msg)
	buf.fields = append(buf.fields, fieldsMap)
	if len(buf.messages) >= 100 {
		l.mu.Unlock()
		l.Flush()
		return
	}
	l.mu.Unlock()
}

func (l *EngineRPCLogger) Debug(msg string, fields ...core.Field) {
	l.addLog(core.LogLevelDebug, msg, fields...)
}

func (l *EngineRPCLogger) Info(msg string, fields ...core.Field) {
	l.addLog(core.LogLevelInfo, msg, fields...)
}

func (l *EngineRPCLogger) Warn(msg string, fields ...core.Field) {
	l.addLog(core.LogLevelWarn, msg, fields...)
}

func (l *EngineRPCLogger) Error(msg string, fields ...core.Field) {
	l.addLog(core.LogLevelError, msg, fields...)
}

func (l *EngineRPCLogger) LogBatch(level core.LogLevel, messages []string, fields []map[string]interface{}) {
	if l == nil || l.engine == nil {
		return
	}
	messageCopy := append([]string(nil), messages...)
	fieldCopy := make([]map[string]interface{}, len(fields))
	for i, values := range fields {
		fieldCopy[i] = cloneLogFields(values)
	}
	l.engine.LogBatch(level, messageCopy, fieldCopy)
}

// Close stops the background flusher and synchronously sends the final batch.
// It is intentionally not part of the public Logger interface so the v1 API
// remains source-compatible while runtime contexts can still own its lifecycle.
func (l *EngineRPCLogger) Close() error {
	if l == nil {
		return nil
	}
	l.closeOnce.Do(func() {
		l.mu.Lock()
		l.closed = true
		l.mu.Unlock()
		close(l.stopCh)
		<-l.doneCh
	})
	return nil
}

func cloneLogFields(values map[string]interface{}) map[string]interface{} {
	if values == nil {
		return nil
	}
	cloned := make(map[string]interface{}, len(values))
	for key, value := range values {
		cloned[key] = cloneLogValue(value)
	}
	return cloned
}

func cloneLogValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case []byte:
		cloned := make([]byte, len(typed))
		copy(cloned, typed)
		return cloned
	case map[string]interface{}:
		return cloneLogFields(typed)
	case []interface{}:
		cloned := make([]interface{}, len(typed))
		for i, item := range typed {
			cloned[i] = cloneLogValue(item)
		}
		return cloned
	default:
		return cloneLogReflect(value)
	}
}

func cloneLogReflect(value interface{}) interface{} {
	if value == nil {
		return nil
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Slice:
		if rv.IsNil() {
			return value
		}
		cloned := reflect.MakeSlice(rv.Type(), rv.Len(), rv.Len())
		for i := 0; i < rv.Len(); i++ {
			item := cloneLogValue(rv.Index(i).Interface())
			cloned.Index(i).Set(reflect.ValueOf(item))
		}
		return cloned.Interface()
	case reflect.Map:
		if rv.IsNil() {
			return value
		}
		cloned := reflect.MakeMapWithSize(rv.Type(), rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			item := cloneLogValue(iter.Value().Interface())
			cloned.SetMapIndex(iter.Key(), reflect.ValueOf(item))
		}
		return cloned.Interface()
	default:
		return value
	}
}
