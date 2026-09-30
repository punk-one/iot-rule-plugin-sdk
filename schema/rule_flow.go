package schema

import "time"

// RuleFlow 定义了规则的完整配置，包含 UI 布局与节点配置
type RuleFlow struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Version         int       `json:"version"`
	Enabled         bool      `json:"enabled"`
	DebugMode       bool      `json:"debug_mode"` // 开启后，节点会发送详细的 Trace 消息到 debug topic
	Description     string    `json:"description"`
	LifecycleStatus string    `json:"lifecycle_status,omitempty"` // 生命周期状态（draft, reviewed, active, archived）
	CreatedAt       time.Time `json:"created_at,omitempty"`
	UpdatedAt       time.Time `json:"updated_at,omitempty"`

	// UI 布局信息 (React Flow 专用，后端不处理但需存储)
	Metadata *FlowMetadata `json:"metadata,omitempty"`

	// 数据源配置
	Source *SourceConfig `json:"source"`

	// 节点列表
	Nodes []*NodeConfig `json:"nodes"`
}

// FlowMetadata 存储前端画布的布局信息
// 使用 map 以支持任意字段（如 source_wires, source_position 等）
type FlowMetadata map[string]interface{}

// Position 表示坐标位置
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// SourceConfig 定义数据源配置
type SourceConfig struct {
	Type       string                 `json:"type"`             // nats, http, mqtt
	Topic      string                 `json:"topic"`            // NATS subject 或 MQTT topic
	DataFormat string                 `json:"data_format"`      // json, protobuf, avro
	Config     map[string]interface{} `json:"config,omitempty"` // 数据源特定配置
}

// NodeConfig 定义单个节点的配置
type NodeConfig struct {
	ID       string                 `json:"id"`                 // 节点唯一ID
	Type     string                 `json:"type"`               // filter, router, map, window, script, sink
	Name     string                 `json:"name"`               // 节点显示名称
	Plugin   string                 `json:"plugin,omitempty"`   // 插件名称（仅用于 plugin 类型节点）
	Config   map[string]interface{} `json:"config"`             // 节点特定配置
	Metadata *NodeMetadata          `json:"metadata,omitempty"` // 节点坐标等UI信息
	Wires    []string               `json:"wires"`              // 下游节点ID列表
	Routes   []RouteConfig          `json:"routes,omitempty"`   // 逻辑路由表
}

// RouteConfig 定义单个逻辑路由
type RouteConfig struct {
	Label string `json:"label"` // 路由标签（EmitTo 使用）
	To    string `json:"to"`    // 目标节点ID
}

// GrayReleaseConfig 灰度发布配置
type GrayReleaseConfig struct {
	RuleID      string            `json:"rule_id"`      // 规则ID
	Version     int               `json:"version"`      // 规则版本
	Enabled     bool              `json:"enabled"`      // 是否启用灰度
	Strategy    string            `json:"strategy"`     // 灰度策略：percentage（比例）| condition（条件）
	Percentage  int               `json:"percentage"`   // 灰度比例（0-100，仅当 strategy=percentage 时有效）
	Condition   map[string]string `json:"condition"`    // 灰度条件（仅当 strategy=condition 时有效，如 {"device_id": "device1"}）
	TargetNodes []string          `json:"target_nodes"` // 目标节点ID列表（灰度流量路由到的节点）
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// NodeMetadata 存储节点的UI元数据
type NodeMetadata struct {
	Position *Position `json:"position,omitempty"`
}

// RuleSnapshot 规则配置快照
// 用于版本管理和不可变归档
type RuleSnapshot struct {
	ID          string    `json:"id"`          // Snapshot ID (UUID)
	RuleID      string    `json:"rule_id"`     // 规则ID
	Version     int       `json:"version"`     // 版本号（从 RuleFlow.Version 继承）
	RuleFlow    RuleFlow  `json:"rule_flow"`   // 完整的规则配置（不可变）
	CreatedAt   time.Time `json:"created_at"`  // 快照创建时间
	CreatedBy   string    `json:"created_by"`  // 创建者（可选）
	Description string    `json:"description"` // 快照描述（可选）
}
