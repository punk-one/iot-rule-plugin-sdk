package sdk

import "github.com/punk-one/iot-rule-plugin-sdk/internal/core"

const EventMetadataTraceIDKey = core.EventMetadataTraceIDKey

type Event = core.Event

func EncodeEvent(e Event) ([]byte, error) {
	return core.EncodeEvent(e)
}

func DecodeEvent(data []byte) (Event, error) {
	return core.DecodeEvent(data)
}

func NewEvent(payload map[string]interface{}, metadata map[string]string) Event {
	return core.NewEvent(payload, metadata)
}

func GetTraceID(e Event) string {
	return core.GetTraceID(e)
}

func SetTraceID(e *Event, traceID string) {
	core.SetTraceID(e, traceID)
}
