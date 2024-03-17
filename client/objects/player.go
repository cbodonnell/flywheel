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
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Player struct {
	ID string
	// TODO: should this be the network manager or the parent game object?
	networkManager *network.NetworkManager
	isLocalPlayer  bool
	State          *gametypes.PlayerState
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
		Timestamp: time.Now().UnixMilli(),
		InputX:    inputX,
		InputY:    inputY,
		InputJump: inputJump,
		DeltaTime: 1.0 / float64(ebiten.TPS()),
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

	// TODO: update the player state with the input
	game.UpdatePlayerState(p.State, cpu)

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
