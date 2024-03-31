package state

import (
	"context"
	"fmt"
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

func (m *InMemoryStateManager) Get(ctx context.Context) (*gametypes.GameState, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	// TODO: assess usage of pointers here and whether values might be better
	copy := &gametypes.GameState{
		Players: make(map[uint32]*gametypes.PlayerState),
	}
	for k, v := range m.gameState.Players {
		copy.Players[k] = &gametypes.PlayerState{
			LastProcessedTimestamp: v.LastProcessedTimestamp,
			Position: gametypes.Position{
				X: v.Position.X,
				Y: v.Position.Y,
			},
			Velocity: gametypes.Velocity{
				X: v.Velocity.X,
				Y: v.Velocity.Y,
			},
			IsOnGround:    v.IsOnGround,
			Animation:     v.Animation,
			AnimationFlip: v.AnimationFlip,
			Object:        v.Object, // TODO: this is a pointer, not a deep copy
		}
	}

	return copy, nil
}

func (m *InMemoryStateManager) Set(ctx context.Context, gameState *gametypes.GameState) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if gameState == nil {
		return fmt.Errorf("game state is nil")
	}

	m.gameState = gameState
	return nil
}
