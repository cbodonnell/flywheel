package state

import (
	"sync"

	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
)

type InMemoryStateManager struct {
	lock      sync.RWMutex
	gameState *gametypes.GameState
}

func NewInMemoryStateManager() *InMemoryStateManager {
	return &InMemoryStateManager{
		gameState: &gametypes.GameState{
			Players: make(map[uint32]*gametypes.PlayerState),
		},
	}
}

func (m *InMemoryStateManager) Get() (*gametypes.GameState, error) {
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

func (m *InMemoryStateManager) Set(gameState *gametypes.GameState) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.gameState = gameState
	return nil
}
