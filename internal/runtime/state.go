package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/punk-one/iot-rule-plugin-sdk/internal/core"
)

type StateManagerImpl struct {
	ruleID  string
	nodeID  string
	manager core.StateStore
}

func NewStateManager(ruleID, nodeID string, manager core.StateStore) core.StateManager {
	return &StateManagerImpl{
		ruleID:  ruleID,
		nodeID:  nodeID,
		manager: manager,
	}
}

func (s *StateManagerImpl) GetState(key string, state interface{}) error {
	return s.manager.GetState(context.Background(), key, state)
}

func (s *StateManagerImpl) SetState(key string, state interface{}) error {
	return s.manager.SetState(context.Background(), key, state)
}

func (s *StateManagerImpl) SetStateWithTTL(key string, state interface{}, ttl time.Duration) error {
	return s.manager.SetStateWithTTL(context.Background(), key, state, ttl)
}

func (s *StateManagerImpl) DeleteState(key string) error {
	return s.manager.DeleteState(context.Background(), key)
}

func (s *StateManagerImpl) GenerateKey(keys []string) string {
	sortedKeys := make([]string, len(keys))
	copy(sortedKeys, keys)
	sort.Strings(sortedKeys)

	keyString := fmt.Sprintf("%s:%s:%s", s.ruleID, s.nodeID, strings.Join(sortedKeys, ":"))
	hash := sha256.Sum256([]byte(keyString))
	return fmt.Sprintf("state:%s", hex.EncodeToString(hash[:]))
}

type rpcStateStore struct {
	engine EngineBridge
}

func NewRPCStateStore(engine EngineBridge) core.StateStore {
	return &rpcStateStore{engine: engine}
}

func (s *rpcStateStore) GetState(ctx context.Context, key string, state interface{}) error {
	if s.engine == nil {
		return fmt.Errorf("state manager not available")
	}
	data, err := s.engine.GetState(key)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return fmt.Errorf("state not found: %s", key)
	}
	return json.Unmarshal(data, state)
}

func (s *rpcStateStore) SetState(ctx context.Context, key string, state interface{}) error {
	return s.SetStateWithTTL(ctx, key, state, 0)
}

func (s *rpcStateStore) SetStateWithTTL(ctx context.Context, key string, state interface{}, ttl time.Duration) error {
	if s.engine == nil {
		return fmt.Errorf("state manager not available")
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return s.engine.SetStateWithTTL(key, data, ttl)
}

func (s *rpcStateStore) DeleteState(ctx context.Context, key string) error {
	if s.engine == nil {
		return fmt.Errorf("state manager not available")
	}
	return s.engine.DeleteState(key)
}
