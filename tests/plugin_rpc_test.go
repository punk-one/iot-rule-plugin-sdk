package sdk_test

import (
	"encoding/json"
	"testing"

	sdk "github.com/punk-one/iot-rule-plugin-sdk"
)

type fallbackProcessor struct {
	sdk.BasePlugin
	count int
}

func (p *fallbackProcessor) Info() sdk.PluginInfo {
	return sdk.PluginInfo{
		ID:      "fallback-processor",
		Name:    "Fallback Processor",
		Version: "1.6.0",
		Type:    sdk.PluginTypeProcessor,
		Capabilities: sdk.PluginCapabilities{
			SupportBatch: false,
		},
	}
}

func (p *fallbackProcessor) OnEvent(e sdk.Event) error {
	p.count++
	return nil
}

func TestPluginRPCServerFallsBackToSingleEventDispatch(t *testing.T) {
	impl := &fallbackProcessor{}
	server := &sdk.PluginRPCServer{Impl: impl}

	events := []sdk.Event{
		sdk.NewEvent(map[string]interface{}{"value": 1}, nil),
		sdk.NewEvent(map[string]interface{}{"value": 2}, nil),
	}
	data, err := json.Marshal(events)
	if err != nil {
		t.Fatalf("marshal events failed: %v", err)
	}

	var reply sdk.ReceiveEventsReply
	if err := server.ReceiveEvents(&sdk.ReceiveEventsArgs{EventsJSON: data}, &reply); err != nil {
		t.Fatalf("ReceiveEvents RPC failed: %v", err)
	}
	if reply.Error != "" {
		t.Fatalf("unexpected plugin error: %s", reply.Error)
	}
	if impl.count != len(events) {
		t.Fatalf("expected %d OnEvent calls, got %d", len(events), impl.count)
	}
}

type batchUnsupportedProcessor struct {
	sdk.BasePlugin
	onEventCount  int
	onEventsCount int
}

func (p *batchUnsupportedProcessor) Info() sdk.PluginInfo {
	return sdk.PluginInfo{
		ID:      "batch-unsupported-processor",
		Name:    "Batch Unsupported Processor",
		Version: "1.6.0",
		Type:    sdk.PluginTypeProcessor,
		Capabilities: sdk.PluginCapabilities{
			SupportBatch: true,
		},
	}
}

func (p *batchUnsupportedProcessor) OnEvent(e sdk.Event) error {
	p.onEventCount++
	return nil
}

func (p *batchUnsupportedProcessor) OnEvents(events []sdk.Event) error {
	p.onEventsCount++
	return sdk.ErrBatchNotSupported
}

func TestPluginRPCServerFallsBackWhenBatchRejectedAtRuntime(t *testing.T) {
	impl := &batchUnsupportedProcessor{}
	server := &sdk.PluginRPCServer{Impl: impl}

	events := []sdk.Event{
		sdk.NewEvent(map[string]interface{}{"value": 1}, nil),
		sdk.NewEvent(map[string]interface{}{"value": 2}, nil),
	}
	data, err := json.Marshal(events)
	if err != nil {
		t.Fatalf("marshal events failed: %v", err)
	}

	var reply sdk.ReceiveEventsReply
	if err := server.ReceiveEvents(&sdk.ReceiveEventsArgs{EventsJSON: data}, &reply); err != nil {
		t.Fatalf("ReceiveEvents RPC failed: %v", err)
	}
	if reply.Error != "" {
		t.Fatalf("unexpected plugin error: %s", reply.Error)
	}
	if impl.onEventsCount != 1 {
		t.Fatalf("expected one OnEvents attempt, got %d", impl.onEventsCount)
	}
	if impl.onEventCount != len(events) {
		t.Fatalf("expected %d OnEvent calls after fallback, got %d", len(events), impl.onEventCount)
	}
}
