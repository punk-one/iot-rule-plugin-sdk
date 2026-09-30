package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/rpc"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/punk-one/iot-rule-plugin-sdk/internal/core"
	"github.com/punk-one/iot-rule-plugin-sdk/internal/runtime"
)

type PluginRPC struct {
	Impl          core.Plugin
	Factory       func() core.Plugin
	defaultHealth core.HealthOptions
}

func (p *PluginRPC) Server(b *plugin.MuxBroker) (interface{}, error) {
	impl := p.Impl
	if p.Factory != nil {
		impl = p.Factory()
	}
	if impl == nil {
		return nil, errors.New("plugin factory returned nil")
	}
	return &PluginRPCServer{
		Impl:          impl,
		broker:        b,
		defaultHealth: p.defaultHealth,
	}, nil
}

func (p *PluginRPC) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &PluginRPCClient{Client: c, Broker: b}, nil
}

type PluginRPCServer struct {
	Impl          core.Plugin
	broker        *plugin.MuxBroker
	ctx           core.RuntimeContext
	defaultHealth core.HealthOptions
}

func (s *PluginRPCServer) Info(args struct{}, reply *InfoReply) error {
	reply.Info = s.Impl.Info()
	return nil
}

func (s *PluginRPCServer) Init(args *InitArgs, reply *InitReply) error {
	reply.Error = core.EncodePluginError(s.Impl.Init(args.Config))
	return nil
}

func (s *PluginRPCServer) Start(args *StartArgs, reply *StartReply) error {
	brokerID := args.BrokerID
	if brokerID == 0 {
		brokerID = 3
	}
	conn, err := s.broker.Dial(brokerID)
	if err != nil {
		reply.Error = fmt.Sprintf("failed to dial Engine RPC connection: %v", err)
		return nil
	}

	rpcClient := rpc.NewClient(conn)
	engineRPC := NewEngineRPCClient(rpcClient)
	ctx := runtime.NewPluginRuntimeContext(
		engineRPC,
		args.RuleID,
		args.NodeID,
		mergeHealthOptions(args.Health, s.defaultHealth),
		nil,
		args.ResourceEventsEnabled,
	)
	s.ctx = ctx

	if err := s.Impl.Start(ctx); err != nil {
		_ = ctx.Health().Unhealthy("SYS_START_FAILED", err.Error(), nil)
		reply.Error = core.EncodePluginError(err)
		return nil
	}
	_ = ctx.Health().Healthy("plugin started", map[string]string{"phase": "start"})
	return nil
}

func (s *PluginRPCServer) ReceiveEvent(args *ReceiveEventArgs, reply *ReceiveEventReply) error {
	var event core.Event
	if err := json.Unmarshal(args.EventJSON, &event); err != nil {
		reply.Error = fmt.Sprintf("failed to unmarshal event: %v", err)
		return nil
	}
	reply.Error = core.EncodePluginError(s.Impl.OnEvent(event))
	return nil
}

func (s *PluginRPCServer) ReceiveEvents(args *ReceiveEventsArgs, reply *ReceiveEventsReply) error {
	var events []core.Event
	if err := json.Unmarshal(args.EventsJSON, &events); err != nil {
		reply.Error = fmt.Sprintf("failed to unmarshal events: %v", err)
		return nil
	}
	reply.Error = core.EncodePluginError(s.dispatchEvents(events))
	return nil
}

func (s *PluginRPCServer) Stop(args struct{}, reply *StopReply) error {
	stopErr := s.Impl.Stop()
	reply.Error = core.EncodePluginError(stopErr)
	if s.ctx != nil {
		if stopErr != nil {
			_ = s.ctx.Health().Unhealthy("SYS_STOP_FAILED", stopErr.Error(), nil)
		} else {
			_ = s.ctx.Health().Report(core.HealthReport{
				Status:  core.HealthStopped,
				Code:    "SYS_STOPPED",
				Message: "plugin stopped",
			})
		}
		if closer, ok := s.ctx.Health().(interface{ Close() error }); ok {
			_ = closer.Close()
		}
		if closer, ok := s.ctx.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	}
	return nil
}

type PluginRPCClient struct {
	Client *rpc.Client
	Broker *plugin.MuxBroker
}

func (c *PluginRPCClient) Info() core.PluginInfo {
	var reply InfoReply
	err := callWithTimeout(c.Client, "Plugin.Info", struct{}{}, &reply, defaultRPCTimeout)
	if err != nil {
		return core.PluginInfo{}
	}
	return reply.Info
}

func (c *PluginRPCClient) Init(cfg core.PluginConfig) error {
	return c.InitWithTimeout(cfg, defaultRPCTimeout)
}

func (c *PluginRPCClient) InitWithTimeout(cfg core.PluginConfig, timeout time.Duration) error {
	args := &InitArgs{Config: cfg}
	var reply InitReply
	if err := callWithTimeout(c.Client, "Plugin.Init", args, &reply, timeout); err != nil {
		return err
	}
	return core.DecodePluginError(reply.Error)
}

func (c *PluginRPCClient) Start(args *StartArgs) error {
	var reply StartReply
	if err := callWithTimeout(c.Client, "Plugin.Start", args, &reply, defaultRPCTimeout); err != nil {
		return err
	}
	return core.DecodePluginError(reply.Error)
}

func (c *PluginRPCClient) ReceiveEvent(e core.Event) error {
	return c.ReceiveEventWithTimeout(e, defaultRPCTimeout)
}

func (c *PluginRPCClient) ReceiveEventWithTimeout(e core.Event, timeout time.Duration) error {
	eventJSON, err := json.Marshal(e)
	if err != nil {
		return err
	}
	args := &ReceiveEventArgs{EventJSON: eventJSON}
	var reply ReceiveEventReply
	if err := callWithTimeout(c.Client, "Plugin.ReceiveEvent", args, &reply, timeout); err != nil {
		return err
	}
	return core.DecodePluginError(reply.Error)
}

func (c *PluginRPCClient) ReceiveEvents(events []core.Event) error {
	return c.ReceiveEventsWithTimeout(events, defaultRPCTimeout)
}

func (c *PluginRPCClient) ReceiveEventsWithTimeout(events []core.Event, timeout time.Duration) error {
	eventsJSON, err := json.Marshal(events)
	if err != nil {
		return err
	}
	args := &ReceiveEventsArgs{EventsJSON: eventsJSON}
	var reply ReceiveEventsReply
	if err := callWithTimeout(c.Client, "Plugin.ReceiveEvents", args, &reply, timeout); err != nil {
		return err
	}
	return core.DecodePluginError(reply.Error)
}

func (c *PluginRPCClient) Stop() error {
	return c.StopWithTimeout(defaultRPCTimeout)
}

func (c *PluginRPCClient) StopWithTimeout(timeout time.Duration) error {
	var reply StopReply
	if err := callWithTimeout(c.Client, "Plugin.Stop", struct{}{}, &reply, timeout); err != nil {
		return err
	}
	return core.DecodePluginError(reply.Error)
}

func (s *PluginRPCServer) dispatchEvents(events []core.Event) error {
	if len(events) == 0 {
		return nil
	}
	if !s.Impl.Info().Capabilities.SupportBatch {
		return s.dispatchEventsSequentially(events)
	}
	if err := s.Impl.OnEvents(events); err != nil {
		if core.IsBatchNotSupported(err) || errors.Is(err, core.ErrNotImplemented) {
			return s.dispatchEventsSequentially(events)
		}
		return err
	}
	return nil
}

func (s *PluginRPCServer) dispatchEventsSequentially(events []core.Event) error {
	for _, event := range events {
		if err := s.Impl.OnEvent(event); err != nil {
			return err
		}
	}
	return nil
}
