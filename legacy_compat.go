package sdk

import transport "github.com/punk-one/iot-rule-plugin-sdk/internal/transport"

// 该文件集中声明 engine / go-plugin / net/rpc 兼容层使用的导出类型。
// 这些类型仍然属于公开 API，但推荐业务插件优先使用更高层的
// `Plugin`、`RuntimeContext`、`Emitter`、`HealthReporter` 等接口。

// LogBatchArgs 是引擎批量日志 RPC 的兼容参数。
type LogBatchArgs = transport.LogBatchArgs

// EmitBatchArgs 是引擎批量事件 RPC 的兼容参数。
type EmitBatchArgs = transport.EmitBatchArgs

// MetricArgs 是引擎指标 RPC 的兼容参数。
type MetricArgs = transport.MetricArgs

// AckArgs 是引擎 Ack RPC 的兼容参数。
type AckArgs = transport.AckArgs

// PublishAckArgs 是引擎业务 Ack 发布 RPC 的兼容参数。
type PublishAckArgs = transport.PublishAckArgs

// StateKeyArgs 是状态查询/删除 RPC 的兼容参数。
type StateKeyArgs = transport.StateKeyArgs

// StateSetArgs 是状态写入 RPC 的兼容参数。
type StateSetArgs = transport.StateSetArgs

// StateSetWithTTLArgs 是带 TTL 状态写入 RPC 的兼容参数。
type StateSetWithTTLArgs = transport.StateSetWithTTLArgs

// StateGetReply 是状态读取 RPC 的兼容返回。
type StateGetReply = transport.StateGetReply

// InfoReply 是插件信息 RPC 的兼容返回。
type InfoReply = transport.InfoReply

// InitArgs / InitReply 是插件初始化 RPC 的兼容参数。
type InitArgs = transport.InitArgs
type InitReply = transport.InitReply

// StartArgs / StartReply 是插件启动 RPC 的兼容参数。
type StartArgs = transport.StartArgs
type StartReply = transport.StartReply

// ReceiveEventArgs / ReceiveEventReply 是单事件下发 RPC 的兼容参数。
type ReceiveEventArgs = transport.ReceiveEventArgs
type ReceiveEventReply = transport.ReceiveEventReply

// ReceiveEventsArgs / ReceiveEventsReply 是批量事件下发 RPC 的兼容参数。
type ReceiveEventsArgs = transport.ReceiveEventsArgs
type ReceiveEventsReply = transport.ReceiveEventsReply

// StopReply 是插件停止 RPC 的兼容返回。
type StopReply = transport.StopReply
