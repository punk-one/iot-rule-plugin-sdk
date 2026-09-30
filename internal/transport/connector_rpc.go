package transport

import (
	"net/rpc"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/punk-one/iot-rule-plugin-sdk/internal/core"
)

type ConnectorRPC struct {
	Impl core.ConnectorPlugin
}

func (p *ConnectorRPC) Server(b *plugin.MuxBroker) (interface{}, error) {
	return &ConnectorRPCServer{Impl: p.Impl}, nil
}

func (p *ConnectorRPC) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ConnectorRPCClient{Client: c}, nil
}

type ConnectorRPCServer struct {
	Impl core.ConnectorPlugin
}

func (s *ConnectorRPCServer) Info(args struct{}, reply *InfoReply) error {
	reply.Info = s.Impl.Info()
	return nil
}

func (s *ConnectorRPCServer) CreateResource(args *ConnectorCreateResourceArgs, reply *ConnectorCreateResourceReply) error {
	handle, err := s.Impl.CreateResource(args.Resource)
	reply.Handle = handle
	reply.Error = core.EncodePluginError(err)
	return nil
}

func (s *ConnectorRPCServer) DestroyResource(args *ConnectorDestroyResourceArgs, reply *ConnectorDestroyResourceReply) error {
	reply.Error = core.EncodePluginError(s.Impl.DestroyResource(args.ProviderHandle))
	return nil
}

func (s *ConnectorRPCServer) Execute(args *ConnectorExecuteArgs, reply *ConnectorExecuteReply) error {
	resp, err := s.Impl.Execute(args.ProviderHandle, args.Request)
	reply.Response = resp
	reply.Error = core.EncodePluginError(err)
	return nil
}

func (s *ConnectorRPCServer) OpenStream(args *ConnectorOpenStreamArgs, reply *ConnectorOpenStreamReply) error {
	resp, err := s.Impl.OpenStream(args.ProviderHandle, args.Request)
	reply.Response = resp
	reply.Error = core.EncodePluginError(err)
	return nil
}

func (s *ConnectorRPCServer) ReceiveStream(args *ConnectorReceiveStreamArgs, reply *ConnectorReceiveStreamReply) error {
	resp, err := s.Impl.ReceiveStream(args.ProviderHandle, args.Request)
	reply.Response = resp
	reply.Error = core.EncodePluginError(err)
	return nil
}

func (s *ConnectorRPCServer) AckStream(args *ConnectorAckStreamArgs, reply *ConnectorAckStreamReply) error {
	reply.Error = core.EncodePluginError(s.Impl.AckStream(args.ProviderHandle, args.Request))
	return nil
}

func (s *ConnectorRPCServer) NackStream(args *ConnectorNackStreamArgs, reply *ConnectorNackStreamReply) error {
	reply.Error = core.EncodePluginError(s.Impl.NackStream(args.ProviderHandle, args.Request))
	return nil
}

func (s *ConnectorRPCServer) GrantCredits(args *ConnectorGrantCreditsArgs, reply *ConnectorGrantCreditsReply) error {
	reply.Error = core.EncodePluginError(s.Impl.GrantCredits(args.ProviderHandle, args.Request))
	return nil
}

func (s *ConnectorRPCServer) CloseStream(args *ConnectorCloseStreamArgs, reply *ConnectorCloseStreamReply) error {
	reply.Error = core.EncodePluginError(s.Impl.CloseStream(args.ProviderHandle, args.Request))
	return nil
}

func (s *ConnectorRPCServer) Probe(args *ConnectorProbeArgs, reply *ConnectorProbeReply) error {
	event, err := s.Impl.Probe(args.ProviderHandle, args.Request)
	reply.Event = event
	reply.Error = core.EncodePluginError(err)
	return nil
}

func (s *ConnectorRPCServer) Stop(args struct{}, reply *StopReply) error {
	reply.Error = core.EncodePluginError(s.Impl.Stop())
	return nil
}

type ConnectorRPCClient struct {
	Client *rpc.Client
}

func (c *ConnectorRPCClient) Info() core.PluginInfo {
	var reply InfoReply
	if err := callWithTimeout(c.Client, "Plugin.Info", struct{}{}, &reply, defaultRPCTimeout); err != nil {
		return core.PluginInfo{}
	}
	return reply.Info
}

func (c *ConnectorRPCClient) CreateResource(resource core.ConnectorResource) (string, error) {
	var reply ConnectorCreateResourceReply
	if err := callWithTimeout(c.Client, "Plugin.CreateResource", &ConnectorCreateResourceArgs{Resource: resource}, &reply, defaultRPCTimeout); err != nil {
		return "", err
	}
	return reply.Handle, core.DecodePluginError(reply.Error)
}

func (c *ConnectorRPCClient) DestroyResource(providerHandle string) error {
	var reply ConnectorDestroyResourceReply
	if err := callWithTimeout(c.Client, "Plugin.DestroyResource", &ConnectorDestroyResourceArgs{ProviderHandle: providerHandle}, &reply, defaultRPCTimeout); err != nil {
		return err
	}
	return core.DecodePluginError(reply.Error)
}

func (c *ConnectorRPCClient) Execute(providerHandle string, req core.ConnectorRequest) (core.ConnectorResponse, error) {
	return c.ExecuteWithTimeout(providerHandle, req, defaultRPCTimeout)
}

func (c *ConnectorRPCClient) ExecuteWithTimeout(providerHandle string, req core.ConnectorRequest, timeout time.Duration) (core.ConnectorResponse, error) {
	var reply ConnectorExecuteReply
	if err := callWithTimeout(c.Client, "Plugin.Execute", &ConnectorExecuteArgs{ProviderHandle: providerHandle, Request: req}, &reply, timeout); err != nil {
		return core.ConnectorResponse{}, err
	}
	return reply.Response, core.DecodePluginError(reply.Error)
}

func (c *ConnectorRPCClient) OpenStream(providerHandle string, req core.StreamOpenRequest) (core.StreamOpenResponse, error) {
	var reply ConnectorOpenStreamReply
	if err := callWithTimeout(c.Client, "Plugin.OpenStream", &ConnectorOpenStreamArgs{ProviderHandle: providerHandle, Request: req}, &reply, defaultRPCTimeout); err != nil {
		return core.StreamOpenResponse{}, err
	}
	return reply.Response, core.DecodePluginError(reply.Error)
}

func (c *ConnectorRPCClient) ReceiveStream(providerHandle string, req core.StreamReceiveRequest) (core.StreamReceiveResponse, error) {
	timeout := defaultRPCTimeout
	if req.WaitTimeoutMS > 0 {
		timeout += time.Duration(req.WaitTimeoutMS) * time.Millisecond
	}
	var reply ConnectorReceiveStreamReply
	if err := callWithTimeout(c.Client, "Plugin.ReceiveStream", &ConnectorReceiveStreamArgs{ProviderHandle: providerHandle, Request: req}, &reply, timeout); err != nil {
		return core.StreamReceiveResponse{}, err
	}
	return reply.Response, core.DecodePluginError(reply.Error)
}

func (c *ConnectorRPCClient) AckStream(providerHandle string, req core.StreamAckRequest) error {
	var reply ConnectorAckStreamReply
	if err := callWithTimeout(c.Client, "Plugin.AckStream", &ConnectorAckStreamArgs{ProviderHandle: providerHandle, Request: req}, &reply, defaultRPCTimeout); err != nil {
		return err
	}
	return core.DecodePluginError(reply.Error)
}

func (c *ConnectorRPCClient) NackStream(providerHandle string, req core.StreamNackRequest) error {
	var reply ConnectorNackStreamReply
	if err := callWithTimeout(c.Client, "Plugin.NackStream", &ConnectorNackStreamArgs{ProviderHandle: providerHandle, Request: req}, &reply, defaultRPCTimeout); err != nil {
		return err
	}
	return core.DecodePluginError(reply.Error)
}

func (c *ConnectorRPCClient) GrantCredits(providerHandle string, req core.StreamGrantCreditsRequest) error {
	var reply ConnectorGrantCreditsReply
	if err := callWithTimeout(c.Client, "Plugin.GrantCredits", &ConnectorGrantCreditsArgs{ProviderHandle: providerHandle, Request: req}, &reply, defaultRPCTimeout); err != nil {
		return err
	}
	return core.DecodePluginError(reply.Error)
}

func (c *ConnectorRPCClient) CloseStream(providerHandle string, req core.StreamCloseRequest) error {
	var reply ConnectorCloseStreamReply
	if err := callWithTimeout(c.Client, "Plugin.CloseStream", &ConnectorCloseStreamArgs{ProviderHandle: providerHandle, Request: req}, &reply, defaultRPCTimeout); err != nil {
		return err
	}
	return core.DecodePluginError(reply.Error)
}

func (c *ConnectorRPCClient) Probe(providerHandle string, req core.ConnectorRequest) (core.ResourceStatusEvent, error) {
	var reply ConnectorProbeReply
	if err := callWithTimeout(c.Client, "Plugin.Probe", &ConnectorProbeArgs{ProviderHandle: providerHandle, Request: req}, &reply, defaultRPCTimeout); err != nil {
		return core.ResourceStatusEvent{}, err
	}
	return reply.Event, core.DecodePluginError(reply.Error)
}

func (c *ConnectorRPCClient) Stop() error {
	var reply StopReply
	if err := callWithTimeout(c.Client, "Plugin.Stop", struct{}{}, &reply, defaultRPCTimeout); err != nil {
		return err
	}
	return core.DecodePluginError(reply.Error)
}
