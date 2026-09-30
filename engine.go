package sdk

import (
	"net/rpc"

	internalruntime "github.com/punk-one/iot-rule-plugin-sdk/internal/runtime"
	transport "github.com/punk-one/iot-rule-plugin-sdk/internal/transport"
)

// EngineRPC 引擎端 RPC 接口
// 插件通过此接口调用引擎能力：发布事件、记录日志、指标、Ack
type EngineRPC = internalruntime.EngineBridge

// EngineRPCClient 引擎端 RPC 客户端包装器
type EngineRPCClient = transport.EngineRPCClient

func NewEngineRPCClient(client *rpc.Client) EngineRPC {
	return transport.NewEngineRPCClient(client)
}

// EdgeSubject 返回边连接使用的 subject 名称。
func EdgeSubject(ruleID, fromNodeID, toNodeID string) string {
	return transport.EdgeSubject(ruleID, fromNodeID, toNodeID)
}
