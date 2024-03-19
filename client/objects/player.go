package objects

import (
	"encoding/json"
	"fmt"
	"image/color"
	"time"

	"github.com/cbodonnell/flywheel/client/network"
	"github.com/cbodonnell/flywheel/pkg/game"
	"github.com/cbodonnell/flywheel/pkg/game/constants"
	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	// MaxPreviousStates is the maximum number of past states to keep
	MaxPreviousStates = 60
)

type Player struct {
	ID string
	// TODO: should this be the network manager or the parent game object?
	networkManager *network.NetworkManager
	isLocalPlayer  bool
	// TODO: make this private with a getter and setter
	State          *gametypes.PlayerState
	previousStates []PreviousState
	pastUpdates    []*messages.ClientPlayerUpdate
}

type PreviousState struct {
	Timestamp int64
	State     *gametypes.PlayerState
}

func NewPlayer(id string, networkManager *network.NetworkManager, state *gametypes.PlayerState) (*Player, error) {
	if networkManager == nil {
		return nil, fmt.Errorf("network manager is required")
	}

	if networkManager.ClientID() == 0 {
		return nil, fmt.Errorf("client ID is required")
	}

	return &Player{
		ID:             id,
		networkManager: networkManager,
		isLocalPlayer:  id == fmt.Sprintf("player-%d", networkManager.ClientID()),
		State:          state,
	}, nil
}

func (p *Player) Update() error {
	if !p.isLocalPlayer {
		return nil
	}

	inputX := 0.0
	rightPressed := ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD)
	leftPressed := ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA)
	if rightPressed && !leftPressed {
		inputX = 1.0
	} else if leftPressed && !rightPressed {
		inputX = -1.0
	}

	inputY := 0.0
	upPressed := ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW)
	downPressed := ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS)
	if upPressed && !downPressed {
		inputY = -1.0
	} else if downPressed && !upPressed {
		inputY = 1.0
	}

	inputJump := ebiten.IsKeyPressed(ebiten.KeySpace)

	cpu := &messages.ClientPlayerUpdate{
		Timestamp:   time.Now().UnixMilli(),
		InputX:      inputX,
		InputY:      inputY,
		InputJump:   inputJump,
		DeltaTime:   1.0 / float64(ebiten.TPS()),
		PastUpdates: p.pastUpdates,
	}
	payload, err := json.Marshal(cpu)
	if err != nil {
		return fmt.Errorf("failed to marshal client player update: %v", err)
	}

	msg := &messages.Message{
		ClientID: p.networkManager.ClientID(),
		Type:     messages.MessageTypeClientPlayerUpdate,
		Payload:  payload,
	}

	if err := p.networkManager.SendUnreliableMessage(msg); err != nil {
		return fmt.Errorf("failed to send client player update: %v", err)
	}

	cpu.PastUpdates = nil // clear the past updates after sending the message
	p.pastUpdates = append(p.pastUpdates, cpu)
	for len(p.pastUpdates) > messages.MaxPreviousUpdates {
		p.pastUpdates = p.pastUpdates[1:]
	}

	game.ApplyInput(p.State, cpu)

	p.previousStates = append(p.previousStates, PreviousState{
		Timestamp: cpu.Timestamp,
		State:     p.State.Copy(), // store a copy as the state will be modified by the game loop
	})
	for len(p.previousStates) > MaxPreviousStates {
		p.previousStates = p.previousStates[1:]
	}

	return nil
}

func (p *Player) Draw(screen *ebiten.Image) {
	playerObject := p.State.Object
	var playerColor color.RGBA
	if p.isLocalPlayer {
		// red if not on ground, blue if on ground
		playerColor = color.RGBA{255, 0, 0, 255} // Red
		if p.State.IsOnGround {
			playerColor = color.RGBA{0, 0, 255, 255} // Blue
		}
	} else {
		playerColor = color.RGBA{0, 255, 60, 255} // Green
		if p.State.IsOnGround {
			playerColor = color.RGBA{200, 0, 200, 255} // Purple
		}
	}
	vector.DrawFilledRect(screen, float32(playerObject.Position.X), float32(float64(screen.Bounds().Dy())-constants.PlayerHeight)-float32(playerObject.Position.Y), float32(playerObject.Size.X), float32(playerObject.Size.Y), playerColor, false)
}

// ReconcileState reconciles the player state with the server state
// by going back through the past client states and checking if it
// the client state for that timestamp matches the server state.
// If it doesn't match, the server state is applied and all of the
// past updates that are after the last processed timestamp are replayed.
func (p *Player) ReconcileState(state *gametypes.PlayerState) {
	foundPreviousState := false
	for i := len(p.previousStates) - 1; i >= 0; i-- {
		ps := p.previousStates[i]
		if ps.Timestamp == state.LastProcessedTimestamp {
			foundPreviousState = true
			if !ps.State.Equal(state) {
				log.Warn("Reconciling player state at timestamp %d for %s", state.LastProcessedTimestamp, p.ID)
				// apply the server state
				p.State.Position.X = state.Position.X
				p.State.Position.Y = state.Position.Y
				p.State.Velocity.X = state.Velocity.X
				p.State.Velocity.Y = state.Velocity.Y
				p.State.IsOnGround = state.IsOnGround
				p.State.Object.Position.X = state.Position.X
				p.State.Object.Position.Y = state.Position.Y

				// process all of the past updates that are after the last processed timestamp
				for j := 0; j < len(p.pastUpdates); j++ {
					if p.pastUpdates[j].Timestamp > state.LastProcessedTimestamp {
						game.ApplyInput(p.State, p.pastUpdates[j])
					}
				}
			}
			break
		}
	}

	if !foundPreviousState {
		log.Warn("Failed to find previous state at timestamp %d for %s for reconciliation", state.LastProcessedTimestamp, p.ID)
	}
}
