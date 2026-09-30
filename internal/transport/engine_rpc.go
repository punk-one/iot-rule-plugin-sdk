package transport

import (
	"errors"
	"fmt"
	"net/rpc"
	"time"

	"github.com/punk-one/iot-rule-plugin-sdk/internal/core"
	"github.com/punk-one/iot-rule-plugin-sdk/internal/runtime"
)

type EngineRPCClient struct {
	client *rpc.Client
}

func NewEngineRPCClient(client *rpc.Client) *EngineRPCClient {
	return &EngineRPCClient{client: client}
}

func (c *EngineRPCClient) Emit(e core.Event) error {
	return c.EmitWithTargets(e, nil)
}

func (c *EngineRPCClient) EmitWithTargets(e core.Event, toNodeIDs []string) error {
	var reply struct{}
	eventJSON, err := core.EncodeEvent(e)
	if err != nil {
		return err
	}
	return callWithTimeout(c.client, "Engine.EmitRPC", struct {
		EventJSON []byte
		ToNodeIDs []string
	}{EventJSON: eventJSON, ToNodeIDs: toNodeIDs}, &reply, defaultRPCTimeout)
}

func (c *EngineRPCClient) Log(level core.LogLevel, msg string, fields map[string]interface{}) {
	var reply struct{}
	_ = callWithTimeout(c.client, "Engine.LogRPC", struct {
		Level  core.LogLevel
		Msg    string
		Fields map[string]interface{}
	}{Level: level, Msg: msg, Fields: fields}, &reply, telemetryRPCTimeout)
}

func (c *EngineRPCClient) LogBatch(level core.LogLevel, messages []string, fields []map[string]interface{}) {
	var reply struct{}
	_ = callWithTimeout(c.client, "Engine.LogBatchRPC", &LogBatchArgs{
		Level:    level,
		Messages: messages,
		Fields:   fields,
	}, &reply, telemetryRPCTimeout)
}

func (c *EngineRPCClient) IncCounter(name string, labels map[string]string) {
	var reply struct{}
	_ = callWithTimeout(c.client, "Engine.IncCounterRPC", &MetricArgs{Name: name, Labels: labels}, &reply, telemetryRPCTimeout)
}

func (c *EngineRPCClient) Observe(name string, value float64, labels map[string]string) {
	var reply struct{}
	_ = callWithTimeout(c.client, "Engine.ObserveRPC", &MetricArgs{Name: name, Value: value, Labels: labels}, &reply, telemetryRPCTimeout)
}

func (c *EngineRPCClient) ReportHealth(args core.ReportHealthArgs) error {
	var reply struct{}
	return callWithTimeout(c.client, "Engine.ReportHealthRPC", &args, &reply, telemetryRPCTimeout)
}

func (c *EngineRPCClient) Ack(eventID string) error {
	var reply struct{}
	return callWithTimeout(c.client, "Engine.AckRPC", &AckArgs{EventID: eventID}, &reply, defaultRPCTimeout)
}

func (c *EngineRPCClient) PublishAck(ack core.AckMessage) error {
	var reply struct{}
	return callWithTimeout(c.client, "Engine.PublishAckRPC", &PublishAckArgs{Ack: ack}, &reply, defaultRPCTimeout)
}

func (c *EngineRPCClient) EmitBatch(events []core.Event) error {
	var reply struct{}
	eventsJSON := make([][]byte, len(events))
	for index, event := range events {
		data, err := core.EncodeEvent(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event %d: %w", index, err)
		}
		eventsJSON[index] = data
	}
	return callWithTimeout(c.client, "Engine.EmitBatchRPC", &EmitBatchArgs{EventsJSON: eventsJSON}, &reply, defaultRPCTimeout)
}

func (c *EngineRPCClient) GetState(key string) ([]byte, error) {
	var reply StateGetReply
	if err := callWithTimeout(c.client, "Engine.GetStateRPC", &StateKeyArgs{Key: key}, &reply, defaultRPCTimeout); err != nil {
		return nil, err
	}
	if !reply.Found {
		return nil, nil
	}
	return reply.Value, nil
}

func (c *EngineRPCClient) SetState(key string, value []byte) error {
	var reply struct{}
	return callWithTimeout(c.client, "Engine.SetStateRPC", &StateSetArgs{Key: key, Value: value}, &reply, defaultRPCTimeout)
}

func (c *EngineRPCClient) SetStateWithTTL(key string, value []byte, ttl time.Duration) error {
	var reply struct{}
	return callWithTimeout(c.client, "Engine.SetStateWithTTLRPC", &StateSetWithTTLArgs{
		Key:      key,
		Value:    value,
		TTLNanos: ttl.Nanoseconds(),
	}, &reply, defaultRPCTimeout)
}

func (c *EngineRPCClient) DeleteState(key string) error {
	var reply struct{}
	return callWithTimeout(c.client, "Engine.DeleteStateRPC", &StateKeyArgs{Key: key}, &reply, defaultRPCTimeout)
}

func (c *EngineRPCClient) ExecuteConnector(req core.ConnectorRequest) (core.ConnectorResponse, error) {
	var reply ConnectorExecuteReply
	if err := callWithTimeout(c.client, "Engine.ConnectorExecuteRPC", &req, &reply, connectorRequestTimeout(req.TimeoutMS)); err != nil {
		return core.ConnectorResponse{}, err
	}
	if reply.Error != "" {
		return core.ConnectorResponse{}, errors.New(reply.Error)
	}
	return reply.Response, nil
}

func (c *EngineRPCClient) OpenConnectorStream(req core.StreamOpenRequest) (core.StreamOpenResponse, error) {
	var reply ConnectorOpenStreamReply
	if err := callWithTimeout(c.client, "Engine.ConnectorOpenStreamRPC", &req, &reply, defaultRPCTimeout); err != nil {
		return core.StreamOpenResponse{}, err
	}
	if reply.Error != "" {
		return core.StreamOpenResponse{}, errors.New(reply.Error)
	}
	return reply.Response, nil
}

func (c *EngineRPCClient) ReceiveConnectorStream(req core.StreamReceiveRequest) (core.StreamReceiveResponse, error) {
	var reply ConnectorReceiveStreamReply
	if err := callWithTimeout(c.client, "Engine.ConnectorReceiveStreamRPC", &req, &reply, connectorReceiveTimeout(req.WaitTimeoutMS)); err != nil {
		return core.StreamReceiveResponse{}, err
	}
	if reply.Error != "" {
		return core.StreamReceiveResponse{}, errors.New(reply.Error)
	}
	return reply.Response, nil
}

func (c *EngineRPCClient) AckConnectorStream(req core.StreamAckRequest) error {
	var reply ConnectorAckStreamReply
	if err := callWithTimeout(c.client, "Engine.ConnectorAckStreamRPC", &req, &reply, defaultRPCTimeout); err != nil {
		return err
	}
	if reply.Error != "" {
		return errors.New(reply.Error)
	}
	return nil
}

func (c *EngineRPCClient) NackConnectorStream(req core.StreamNackRequest) error {
	var reply ConnectorNackStreamReply
	if err := callWithTimeout(c.client, "Engine.ConnectorNackStreamRPC", &req, &reply, defaultRPCTimeout); err != nil {
		return err
	}
	if reply.Error != "" {
		return errors.New(reply.Error)
	}
	return nil
}

func (c *EngineRPCClient) GrantConnectorStream(req core.StreamGrantCreditsRequest) error {
	var reply ConnectorGrantCreditsReply
	if err := callWithTimeout(c.client, "Engine.ConnectorGrantCreditsRPC", &req, &reply, defaultRPCTimeout); err != nil {
		return err
	}
	if reply.Error != "" {
		return errors.New(reply.Error)
	}
	return nil
}

func (c *EngineRPCClient) CloseConnectorStream(req core.StreamCloseRequest) error {
	var reply ConnectorCloseStreamReply
	if err := callWithTimeout(c.client, "Engine.ConnectorCloseStreamRPC", &req, &reply, defaultRPCTimeout); err != nil {
		return err
	}
	if reply.Error != "" {
		return errors.New(reply.Error)
	}
	return nil
}

func (c *EngineRPCClient) CurrentResourceStatus(resourceRef string) (core.ResourceStatusEvent, bool) {
	var reply CurrentResourceStatusReply
	if err := callWithTimeout(c.client, "Engine.CurrentResourceStatusRPC", &CurrentResourceStatusArgs{
		ResourceRef: resourceRef,
	}, &reply, defaultRPCTimeout); err != nil {
		return core.ResourceStatusEvent{}, false
	}
	return reply.Event, reply.Found
}

func (c *EngineRPCClient) NextResourceEvent(timeout time.Duration) (core.ResourceStatusEvent, bool, error) {
	var reply NextResourceEventReply
	if err := callWithTimeout(c.client, "Engine.NextResourceEventRPC", &NextResourceEventArgs{
		TimeoutMS: int(timeout / time.Millisecond),
	}, &reply, connectorReceiveTimeout(int(timeout/time.Millisecond))); err != nil {
		return core.ResourceStatusEvent{}, false, err
	}
	return reply.Event, reply.OK, nil
}

func connectorRequestTimeout(timeoutMS int) time.Duration {
	if timeoutMS <= 0 {
		return defaultRPCTimeout
	}
	return time.Duration(timeoutMS)*time.Millisecond + time.Second
}

func connectorReceiveTimeout(waitTimeoutMS int) time.Duration {
	if waitTimeoutMS <= 0 {
		return defaultRPCTimeout
	}
	return time.Duration(waitTimeoutMS)*time.Millisecond + time.Second
}

var _ runtime.EngineBridge = (*EngineRPCClient)(nil)

func EdgeSubject(ruleID, fromNodeID, toNodeID string) string {
	return fmt.Sprintf("rule.edge.%s.%s.%s", ruleID, fromNodeID, toNodeID)
}
