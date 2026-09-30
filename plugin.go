package sdk

import (
	"github.com/hashicorp/go-plugin"
	"github.com/punk-one/iot-rule-plugin-sdk/internal/core"
	transport "github.com/punk-one/iot-rule-plugin-sdk/internal/transport"
)

type Plugin = core.Plugin
type BasePlugin = core.BasePlugin

type PluginType = core.PluginType

const (
	PluginTypeSource    = core.PluginTypeSource
	PluginTypeProcessor = core.PluginTypeProcessor
	PluginTypeSink      = core.PluginTypeSink
	PluginTypeUtility   = core.PluginTypeUtility
	PluginTypeConnector = core.PluginTypeConnector
)

type PluginInfo = core.PluginInfo
type PluginCapabilities = core.PluginCapabilities
type ConnectorDescriptor = core.ConnectorDescriptor
type ConnectorBindingSpec = core.ConnectorBindingSpec
type PluginConfig = core.PluginConfig
type ConnectorClient = core.ConnectorClient
type StreamConsumer = core.StreamConsumer
type ConnectorResource = core.ConnectorResource
type ConnectorRequest = core.ConnectorRequest
type ConnectorRequestKind = core.ConnectorRequestKind
type ConnectorResponse = core.ConnectorResponse
type StreamDeliverySemantics = core.StreamDeliverySemantics
type StreamOpenRequest = core.StreamOpenRequest
type StreamOpenResponse = core.StreamOpenResponse
type StreamReceiveRequest = core.StreamReceiveRequest
type StreamReceiveResponse = core.StreamReceiveResponse
type StreamMessage = core.StreamMessage
type StreamAckRequest = core.StreamAckRequest
type StreamNackRequest = core.StreamNackRequest
type StreamGrantCreditsRequest = core.StreamGrantCreditsRequest
type StreamCloseRequest = core.StreamCloseRequest
type HealthPolicy = core.HealthPolicy
type ProviderPolicy = core.ProviderPolicy
type QuotaPolicy = core.QuotaPolicy
type ResourceStatus = core.ResourceStatus
type ResourceStatusEvent = core.ResourceStatusEvent
type WarmupPolicy = core.WarmupPolicy

const (
	WarmupPolicyLazy     = core.WarmupPolicyLazy
	WarmupPolicyPreload  = core.WarmupPolicyPreload
	WarmupPolicyAlwaysOn = core.WarmupPolicyAlwaysOn

	ConnectorSessionPolicyPerOperation   = core.ConnectorSessionPolicyPerOperation
	ConnectorSessionPolicyScheduledReuse = core.ConnectorSessionPolicyScheduledReuse
	ConnectorSessionPolicyPersistent     = core.ConnectorSessionPolicyPersistent
	ConnectorSessionPolicyPooled         = core.ConnectorSessionPolicyPooled
	ConnectorSessionPolicyStateless      = core.ConnectorSessionPolicyStateless
	ConnectorSessionPolicyIdleTimeout    = core.ConnectorSessionPolicyIdleTimeout

	ResourceStatusUnknown      = core.ResourceStatusUnknown
	ResourceStatusHealthy      = core.ResourceStatusHealthy
	ResourceStatusDegraded     = core.ResourceStatusDegraded
	ResourceStatusCongested    = core.ResourceStatusCongested
	ResourceStatusDisconnected = core.ResourceStatusDisconnected
	ResourceStatusReadonly     = core.ResourceStatusReadonly
	ResourceStatusBlocked      = core.ResourceStatusBlocked
	ResourceStatusRecovering   = core.ResourceStatusRecovering

	ConnectorRequestRead       = core.ConnectorRequestRead
	ConnectorRequestWrite      = core.ConnectorRequestWrite
	ConnectorRequestBatchRead  = core.ConnectorRequestBatchRead
	ConnectorRequestBatchWrite = core.ConnectorRequestBatchWrite
	ConnectorRequestMetadata   = core.ConnectorRequestMetadata

	StreamDeliveryBestEffort  = core.StreamDeliveryBestEffort
	StreamDeliveryAtLeastOnce = core.StreamDeliveryAtLeastOnce
)

type LogLevel = core.LogLevel

const (
	LogLevelDebug = core.LogLevelDebug
	LogLevelInfo  = core.LogLevelInfo
	LogLevelWarn  = core.LogLevelWarn
	LogLevelError = core.LogLevelError
)

type Field = core.Field

// PluginRPC 实现了 go-plugin 的 Plugin 接口
type PluginRPC = transport.PluginRPC

// PluginRPCServer RPC 服务端
type PluginRPCServer = transport.PluginRPCServer

// PluginRPCClient RPC 客户端
type PluginRPCClient = transport.PluginRPCClient

// HandshakeConfig 是 go-plugin 的握手配置
// 确保插件和引擎使用的协议版本一致
// Deprecated: 仅为 v1 源码兼容保留。修改此变量不会改变 SDK 内部握手；
// 引擎加载插件时应调用 DefaultHandshakeConfig 获取固定配置的副本。
var HandshakeConfig = transport.DefaultHandshakeConfig()

// DefaultHandshakeConfig 返回不可变 v1 握手配置的独立副本。
func DefaultHandshakeConfig() plugin.HandshakeConfig {
	return transport.DefaultHandshakeConfig()
}
