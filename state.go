package sdk

import (
	"github.com/punk-one/iot-rule-plugin-sdk/internal/core"
	internalruntime "github.com/punk-one/iot-rule-plugin-sdk/internal/runtime"
)

type StateManager = core.StateManager
type StateStore = core.StateStore

// NewStateManager 创建 StateManager
func NewStateManager(ruleID, nodeID string, manager StateStore) StateManager {
	return internalruntime.NewStateManager(ruleID, nodeID, manager)
}

func newRPCStateStore(engine EngineRPC) StateStore {
	return internalruntime.NewRPCStateStore(engine)
}
