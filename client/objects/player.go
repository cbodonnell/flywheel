package objects

import (
	"encoding/json"
	"fmt"
	"image/color"
	"time"

	"github.com/cbodonnell/flywheel/client/input"
	"github.com/cbodonnell/flywheel/client/network"
	"github.com/cbodonnell/flywheel/pkg/game"
	"github.com/cbodonnell/flywheel/pkg/game/constants"
	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/solarlune/resolv"
)

const (
	// MaxPreviousStates is the maximum number of past states to keep
	MaxPreviousStates = 60
)

type Player struct {
	BaseObject

	ID             string
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

	state.Object = resolv.NewObject(state.Position.X, state.Position.Y, constants.PlayerWidth, constants.PlayerHeight, game.CollisionSpaceTagPlayer)

	return &Player{
		BaseObject: BaseObject{
			Children: make(map[string]GameObject),
		},
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
	rightPressed := input.IsRightPressed()
	leftPressed := input.IsLeftPressed()
	if rightPressed && !leftPressed {
		inputX = 1.0
	} else if leftPressed && !rightPressed {
		inputX = -1.0
	}

	inputY := 0.0
	upPressed := input.IsUpPressed()
	downPressed := input.IsDownPressed()
	if upPressed && !downPressed {
		inputY = -1.0
	} else if downPressed && !upPressed {
		inputY = 1.0
	}

	inputJump := input.IsJumpPressed()

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
	vector.DrawFilledRect(screen, float32(playerObject.Position.X), float32(float64(screen.Bounds().Dy())-playerObject.Size.Y)-float32(playerObject.Position.Y), float32(playerObject.Size.X), float32(playerObject.Size.Y), playerColor, false)
}

func (p *Player) InterpolateState(from *gametypes.PlayerState, to *gametypes.PlayerState, factor float64) {
	p.State.LastProcessedTimestamp = to.LastProcessedTimestamp
	p.State.Position.X = from.Position.X + (to.Position.X-from.Position.X)*factor
	p.State.Position.Y = from.Position.Y + (to.Position.Y-from.Position.Y)*factor
	p.State.Velocity.X = to.Velocity.X
	p.State.Velocity.Y = to.Velocity.X
	p.State.IsOnGround = to.IsOnGround
	p.State.Object.Position.X = p.State.Position.X
	p.State.Object.Position.Y = p.State.Position.Y
}

func (p *Player) ExtrapolateState(from *gametypes.PlayerState, to *gametypes.PlayerState, factor float64) {
	p.State.LastProcessedTimestamp = to.LastProcessedTimestamp
	p.State.Position.X = to.Position.X + (to.Position.X-from.Position.X)*factor
	p.State.Position.Y = to.Position.Y + (to.Position.Y-from.Position.Y)*factor
	p.State.Velocity.X = to.Velocity.X
	p.State.Velocity.Y = to.Velocity.Y
	p.State.IsOnGround = to.IsOnGround
	p.State.Object.Position.X = p.State.Position.X
	p.State.Object.Position.Y = p.State.Position.Y
}

// ReconcileState reconciles the player state with the server state
// by going back through the past client states and checking if it
// the client state for that timestamp matches the server state.
// If it doesn't match, the server state is applied and all of the
// past updates that are after the last processed timestamp are replayed.
func (p *Player) ReconcileState(state *gametypes.PlayerState) error {
	if state.LastProcessedTimestamp == 0 {
		// no previous state to reconcile with
		return nil
	}

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
		return fmt.Errorf("failed to find previous state for timestamp %d", state.LastProcessedTimestamp)
	}

	return nil
}
