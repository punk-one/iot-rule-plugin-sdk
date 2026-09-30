package core

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync/atomic"
	"time"
)

const EventMetadataTraceIDKey = "trace_id"

var lastEventID atomic.Int64

// Event payload values must be JSON-compatible. Clone deep-copies the mutable
// JSON containers plus common typed slices used by telemetry plugins.
type Event struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
	Metadata  map[string]string      `json:"metadata"`

	RuleID string `json:"-"`
	Source string `json:"-"`
}

func EncodeEvent(e Event) ([]byte, error) {
	return json.Marshal(e)
}

func DecodeEvent(data []byte) (Event, error) {
	var e Event
	err := json.Unmarshal(data, &e)
	return e, err
}

func NewEvent(payload map[string]interface{}, metadata map[string]string) Event {
	now := time.Now()
	if metadata == nil {
		metadata = make(map[string]string)
	}
	return Event{
		ID:        fmt.Sprintf("evt_%d", nextEventID(now.UnixNano())),
		Timestamp: now,
		Payload:   payload,
		Metadata:  metadata,
	}
}

func nextEventID(candidate int64) int64 {
	for {
		previous := lastEventID.Load()
		if candidate <= previous {
			candidate = previous + 1
		}
		if lastEventID.CompareAndSwap(previous, candidate) {
			return candidate
		}
	}
}

func (e Event) Clone() Event {
	clone := e
	clone.Payload = clonePayload(e.Payload)
	clone.Metadata = cloneMetadata(e.Metadata)
	return clone
}

func GetTraceID(e Event) string {
	if e.Metadata == nil {
		return ""
	}
	return e.Metadata[EventMetadataTraceIDKey]
}

func SetTraceID(e *Event, traceID string) {
	if e == nil {
		return
	}
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[EventMetadataTraceIDKey] = traceID
}

func cloneMetadata(metadata map[string]string) map[string]string {
	if metadata == nil {
		return nil
	}
	cloned := make(map[string]string, len(metadata))
	for key, value := range metadata {
		cloned[key] = value
	}
	return cloned
}

func clonePayload(payload map[string]interface{}) map[string]interface{} {
	if payload == nil {
		return nil
	}
	cloned := make(map[string]interface{}, len(payload))
	for key, value := range payload {
		cloned[key] = cloneValue(value)
	}
	return cloned
}

func cloneValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case json.RawMessage:
		cloned := make(json.RawMessage, len(typed))
		copy(cloned, typed)
		return cloned
	case []byte:
		return cloneSlice(typed)
	case map[string]interface{}:
		return clonePayload(typed)
	case map[string]string:
		cloned := make(map[string]string, len(typed))
		for key, item := range typed {
			cloned[key] = item
		}
		return cloned
	case []interface{}:
		cloned := make([]interface{}, len(typed))
		for index, item := range typed {
			cloned[index] = cloneValue(item)
		}
		return cloned
	case []string:
		return cloneSlice(typed)
	case []bool:
		return cloneSlice(typed)
	case []int:
		return cloneSlice(typed)
	case []int32:
		return cloneSlice(typed)
	case []int64:
		return cloneSlice(typed)
	case []uint:
		return cloneSlice(typed)
	case []uint32:
		return cloneSlice(typed)
	case []uint64:
		return cloneSlice(typed)
	case []float32:
		return cloneSlice(typed)
	case []float64:
		return cloneSlice(typed)
	default:
		if cloned, ok := cloneReflectContainer(reflect.ValueOf(value)); ok {
			return cloned.Interface()
		}
		return typed
	}
}

func cloneSlice[T any](values []T) []T {
	if values == nil {
		return nil
	}
	cloned := make([]T, len(values))
	copy(cloned, values)
	return cloned
}

// cloneReflectContainer covers named and less-common JSON-compatible maps,
// slices and arrays without putting reflection on the common hot paths above.
func cloneReflectContainer(value reflect.Value) (reflect.Value, bool) {
	if !value.IsValid() {
		return reflect.Value{}, false
	}

	switch value.Kind() {
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type()), true
		}
		cloned := reflect.New(value.Type().Elem())
		cloned.Elem().Set(cloneReflectElement(value.Elem(), value.Type().Elem()))
		return cloned, true
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type()), true
		}
		cloned := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for i := 0; i < value.Len(); i++ {
			cloned.Index(i).Set(cloneReflectElement(value.Index(i), value.Type().Elem()))
		}
		return cloned, true
	case reflect.Array:
		cloned := reflect.New(value.Type()).Elem()
		for i := 0; i < value.Len(); i++ {
			cloned.Index(i).Set(cloneReflectElement(value.Index(i), value.Type().Elem()))
		}
		return cloned, true
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type()), true
		}
		cloned := reflect.MakeMapWithSize(value.Type(), value.Len())
		iter := value.MapRange()
		for iter.Next() {
			cloned.SetMapIndex(iter.Key(), cloneReflectElement(iter.Value(), value.Type().Elem()))
		}
		return cloned, true
	case reflect.Struct:
		// Structs are not part of the documented JSON payload model, but copying
		// their exported mutable fields avoids retaining aliases when a plugin
		// passes a typed DTO. Unexported fields stay value-copied (for example,
		// time.Time remains safe and preserves its internal representation).
		cloned := reflect.New(value.Type()).Elem()
		cloned.Set(value)
		for i := 0; i < value.NumField(); i++ {
			target := cloned.Field(i)
			if !target.CanSet() || !value.Field(i).CanInterface() {
				continue
			}
			target.Set(cloneReflectElement(value.Field(i), target.Type()))
		}
		return cloned, true
	default:
		return reflect.Value{}, false
	}
}

func cloneReflectElement(value reflect.Value, targetType reflect.Type) reflect.Value {
	if !value.IsValid() || !value.CanInterface() {
		return value
	}
	cloned := cloneValue(value.Interface())
	if cloned == nil {
		return reflect.Zero(targetType)
	}
	result := reflect.ValueOf(cloned)
	if result.Type().AssignableTo(targetType) {
		return result
	}
	if result.Type().ConvertibleTo(targetType) {
		return result.Convert(targetType)
	}
	return value
}
