package state

import (
	"sync"

	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
)

type MemoryStateManager struct {
	lock      sync.RWMutex
	gameState *gametypes.GameState
}

func NewMemoryStateManager() *MemoryStateManager {
	return &MemoryStateManager{
		gameState: &gametypes.GameState{
			Players: make(map[uint32]*gametypes.PlayerState),
		},
	}
}

func (m *MemoryStateManager) Get() (*gametypes.GameState, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	copy := &gametypes.GameState{
		Players: make(map[uint32]*gametypes.PlayerState),
	}
	for k, v := range m.gameState.Players {
		copy.Players[k] = &gametypes.PlayerState{
			P: gametypes.Position{
				X: v.P.X,
				Y: v.P.Y,
			},
		}
	}

	return copy, nil
}

func (m *MemoryStateManager) Set(gameState *gametypes.GameState) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.gameState = gameState
	return nil
}
