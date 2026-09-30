package transport

import (
	"sync/atomic"
	"testing"

	"github.com/punk-one/iot-rule-plugin-sdk/internal/core"
)

type factoryPlugin struct {
	core.BasePlugin
	id int64
}

func (p *factoryPlugin) Info() core.PluginInfo {
	return core.PluginInfo{ID: "factory-test"}
}

func TestPluginRPCFactoryCreatesIsolatedImplementations(t *testing.T) {
	var calls atomic.Int64
	rpcPlugin := &PluginRPC{Factory: func() core.Plugin {
		return &factoryPlugin{id: calls.Add(1)}
	}}

	firstRaw, err := rpcPlugin.Server(nil)
	if err != nil {
		t.Fatalf("create first server: %v", err)
	}
	secondRaw, err := rpcPlugin.Server(nil)
	if err != nil {
		t.Fatalf("create second server: %v", err)
	}

	first := firstRaw.(*PluginRPCServer).Impl.(*factoryPlugin)
	second := secondRaw.(*PluginRPCServer).Impl.(*factoryPlugin)
	if first == second || first.id == second.id {
		t.Fatalf("factory returned shared implementation: first=%p second=%p", first, second)
	}
	if calls.Load() != 2 {
		t.Fatalf("factory calls=%d, want 2", calls.Load())
	}
}

func TestPluginRPCFactoryRejectsNilImplementation(t *testing.T) {
	rpcPlugin := &PluginRPC{Factory: func() core.Plugin { return nil }}
	if _, err := rpcPlugin.Server(nil); err == nil {
		t.Fatal("expected nil factory result to fail")
	}
}
