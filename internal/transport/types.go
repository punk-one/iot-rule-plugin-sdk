package transport

import "github.com/punk-one/iot-rule-plugin-sdk/internal/core"

type LogBatchArgs struct {
	Level    core.LogLevel
	Messages []string
	Fields   []map[string]interface{}
}

type EmitBatchArgs struct {
	EventsJSON [][]byte
	ToNodeIDs  []string
}

type MetricArgs struct {
	Name   string
	Value  float64
	Labels map[string]string
}

type AckArgs struct {
	EventID string
}

type PublishAckArgs struct {
	Ack core.AckMessage
}

type StateKeyArgs struct {
	Key string
}

type StateSetArgs struct {
	Key   string
	Value []byte
}

type StateSetWithTTLArgs struct {
	Key      string
	Value    []byte
	TTLNanos int64
}

type StateGetReply struct {
	Value []byte
	Found bool
}

type InfoReply struct{ Info core.PluginInfo }
type InitArgs struct{ Config core.PluginConfig }
type InitReply struct{ Error string }

type StartArgs struct {
	RuleID                string
	NodeID                string
	Health                core.HealthOptions
	ResourceEventsEnabled bool
	BrokerID              uint32
}

type StartReply struct{ Error string }
type ReceiveEventArgs struct{ EventJSON []byte }
type ReceiveEventReply struct{ Error string }
type ReceiveEventsArgs struct{ EventsJSON []byte }
type ReceiveEventsReply struct{ Error string }
type StopReply struct{ Error string }

type CurrentResourceStatusArgs struct {
	ResourceRef string
}

type CurrentResourceStatusReply struct {
	Event core.ResourceStatusEvent
	Found bool
}

type NextResourceEventArgs struct {
	TimeoutMS int
}

type NextResourceEventReply struct {
	Event core.ResourceStatusEvent
	OK    bool
}

type ConnectorCreateResourceArgs struct {
	Resource core.ConnectorResource
}

type ConnectorCreateResourceReply struct {
	Handle string
	Error  string
}

type ConnectorDestroyResourceArgs struct {
	ProviderHandle string
}

type ConnectorDestroyResourceReply struct {
	Error string
}

type ConnectorExecuteArgs struct {
	ProviderHandle string
	Request        core.ConnectorRequest
}

type ConnectorExecuteReply struct {
	Response core.ConnectorResponse
	Error    string
}

type ConnectorOpenStreamArgs struct {
	ProviderHandle string
	Request        core.StreamOpenRequest
}

type ConnectorOpenStreamReply struct {
	Response core.StreamOpenResponse
	Error    string
}

type ConnectorReceiveStreamArgs struct {
	ProviderHandle string
	Request        core.StreamReceiveRequest
}

type ConnectorReceiveStreamReply struct {
	Response core.StreamReceiveResponse
	Error    string
}

type ConnectorAckStreamArgs struct {
	ProviderHandle string
	Request        core.StreamAckRequest
}

type ConnectorAckStreamReply struct {
	Error string
}

type ConnectorNackStreamArgs struct {
	ProviderHandle string
	Request        core.StreamNackRequest
}

type ConnectorNackStreamReply struct {
	Error string
}

type ConnectorGrantCreditsArgs struct {
	ProviderHandle string
	Request        core.StreamGrantCreditsRequest
}

type ConnectorGrantCreditsReply struct {
	Error string
}

type ConnectorCloseStreamArgs struct {
	ProviderHandle string
	Request        core.StreamCloseRequest
}

type ConnectorCloseStreamReply struct {
	Error string
}

type ConnectorProbeArgs struct {
	ProviderHandle string
	Request        core.ConnectorRequest
}

type ConnectorProbeReply struct {
	Event core.ResourceStatusEvent
	Error string
}
