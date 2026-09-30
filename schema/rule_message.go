package schema

// RuleMessage 定义了在 NATS 中传输的统一消息体
type RuleMessage struct {
	// 基础数据
	Payload   []map[string]interface{} `json:"payload"`    // 批量数据
	BatchSize int                      `json:"batch_size"` // 批次大小
	Metadata  map[string]string        `json:"metadata"`   // HTTP Header 或 MQTT UserProperties

	// 追踪与上下文
	ID        string `json:"id"`        // 消息唯一ID (UUID)
	SourceID  string `json:"source_id"` // 数据源ID
	Timestamp int64  `json:"timestamp"` // Unix 时间戳（纳秒）
	TraceID   string `json:"trace_id"`  // OpenTelemetry TraceID

	// 路由控制
	RuleID     string `json:"rule_id"`      // 规则ID
	FromNodeID string `json:"from_node_id"` // 来源节点ID

	// 控制标志
	IsDryRun bool `json:"is_dry_run"` // 如果为 true，Sink 节点只打印日志，不执行写操作
	IsDebug  bool `json:"is_debug"`   // 如果为 true，每个节点都会向 debug topic 发送中间状态

	// 错误处理
	Error string `json:"error,omitempty"`
}
