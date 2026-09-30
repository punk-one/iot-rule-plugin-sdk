package core

import "time"

const (
	AckStatusSuccess = "success"
	AckStatusFailed  = "failed"
)

// AckMessage 表示下游节点对 Buffer 记录的业务确认。
type AckMessage struct {
	RecordID     string    `json:"record_id"`
	RuleID       string    `json:"rule_id,omitempty"`
	BufferNodeID string    `json:"buffer_node_id,omitempty"`
	FromNodeID   string    `json:"from_node_id,omitempty"`
	EventID      string    `json:"event_id,omitempty"`
	Status       string    `json:"status"`
	Error        string    `json:"error,omitempty"`
	AckedAt      time.Time `json:"acked_at"`
}
