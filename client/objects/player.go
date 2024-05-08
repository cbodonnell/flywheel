package objects

import (
	"encoding/json"
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/cbodonnell/flywheel/client/animations"
	"github.com/cbodonnell/flywheel/client/fonts"
	"github.com/cbodonnell/flywheel/client/input"
	"github.com/cbodonnell/flywheel/client/network"
	"github.com/cbodonnell/flywheel/pkg/game/constants"
	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/solarlune/resolv"
	"golang.org/x/image/font"
)

const (
	// MaxPreviousStates is the maximum number of past states to keep
	MaxPreviousStates = 60
)

type Player struct {
	*BaseObject

	ID             string
	networkManager *network.NetworkManager
	debug          bool
	isLocalPlayer  bool
	// TODO: make this private with a getter and setter
	State          *gametypes.PlayerState
	previousStates []PreviousState
	pastUpdates    []*messages.ClientPlayerUpdate

	animations                 map[gametypes.PlayerAnimation]*animations.Animation
	lastDrawnAnimationSequence uint8
}

type PreviousState struct {
	Timestamp int64
	State     *gametypes.PlayerState
}

func NewPlayer(id string, networkManager *network.NetworkManager, state *gametypes.PlayerState) (*Player, error) {

	// TODO: is network manager required for non-local players?
	if networkManager == nil {
		return nil, fmt.Errorf("network manager is required")
	}

	if networkManager.ClientID() == 0 {
		return nil, fmt.Errorf("client ID is required")
	}

	isLocalPlayer := id == fmt.Sprintf("player-%d", networkManager.ClientID())

	state.Object = resolv.NewObject(state.Position.X, state.Position.Y, constants.PlayerWidth, constants.PlayerHeight, gametypes.CollisionSpaceTagPlayer)

	baseObjectOpts := &NewBaseObjectOpts{
		ZIndex: 20,
	}
	if isLocalPlayer {
		baseObjectOpts.ZIndex = 30
	}

	return &Player{
		BaseObject:     NewBaseObject(id, baseObjectOpts),
		ID:             id,
		networkManager: networkManager,
		isLocalPlayer:  isLocalPlayer,
		// debug:          true,
		State: state,
		animations: map[gametypes.PlayerAnimation]*animations.Animation{
			gametypes.PlayerAnimationIdle:    animations.NewPlayerIdleAnimation(),
			gametypes.PlayerAnimationRun:     animations.NewPlayerRunAnimation(),
			gametypes.PlayerAnimationJump:    animations.NewPlayerJumpAnimation(),
			gametypes.PlayerAnimationFall:    animations.NewPlayerFallAnimation(),
			gametypes.PlayerAnimationAttack1: animations.NewPlayerAttack1Animation(),
			gametypes.PlayerAnimationAttack2: animations.NewPlayerAttack2Animation(),
			gametypes.PlayerAnimationAttack3: animations.NewPlayerAttack3Animation(),
			gametypes.PlayerAnimationDead:    animations.NewPlayerDeadAnimation(),
		},
	}, nil
}

func (o *Player) Update() error {
	o.animations[o.State.Animation].Update()

	if !o.isLocalPlayer {
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

	inputJump := input.IsJumpJustPressed()

	inputAttack1 := input.IsAttack1JustPressed()
	inputAttack2 := input.IsAttack2JustPressed()
	inputAttack3 := input.IsAttack3JustPressed()

	inputRespawn := input.IsRespawnJustPressed()

	cpu := &messages.ClientPlayerUpdate{
		Timestamp:    time.Now().UnixMilli(),
		InputX:       inputX,
		InputY:       inputY,
		InputJump:    inputJump,
		InputAttack1: inputAttack1,
		InputAttack2: inputAttack2,
		InputAttack3: inputAttack3,
		InputRespawn: inputRespawn,
		DeltaTime:    1.0 / float64(ebiten.TPS()),
		PastUpdates:  o.pastUpdates,
	}
	payload, err := json.Marshal(cpu)
	if err != nil {
		return fmt.Errorf("failed to marshal client player update: %v", err)
	}

	msg := &messages.Message{
		ClientID: o.networkManager.ClientID(),
		Type:     messages.MessageTypeClientPlayerUpdate,
		Payload:  payload,
	}

	if err := o.networkManager.SendUnreliableMessage(msg); err != nil {
		return fmt.Errorf("failed to send client player update: %v", err)
	}

	cpu.PastUpdates = nil // clear the past updates after sending the message
	o.pastUpdates = append(o.pastUpdates, cpu)
	for len(o.pastUpdates) > messages.MaxPreviousUpdates {
		o.pastUpdates = o.pastUpdates[1:]
	}

	o.State.ApplyInput(cpu)

	o.previousStates = append(o.previousStates, PreviousState{
		Timestamp: cpu.Timestamp,
		State:     o.State.Copy(), // store a copy as the state will be modified by the game loop
	})
	for len(o.previousStates) > MaxPreviousStates {
		o.previousStates = o.previousStates[1:]
	}

	return nil
}

func (o *Player) Draw(screen *ebiten.Image) {
	if o.State.AnimationSequence != o.lastDrawnAnimationSequence {
		o.animations[o.State.Animation].Reset()
	}
	o.animations[o.State.Animation].Draw(screen, o.State.Position.X, o.State.Position.Y, o.State.AnimationFlip)
	o.lastDrawnAnimationSequence = o.State.AnimationSequence

	// Draw Name
	t := strings.ToUpper(o.State.Name)
	f := fonts.TTFSmallFont
	bounds, _ := font.BoundString(f, t)
	op := &ebiten.DrawImageOptions{}
	offsetY := float64(24)
	op.GeoM.Translate(float64(o.State.Position.X)+constants.PlayerWidth/2-float64(bounds.Max.X>>6)/2, float64(screen.Bounds().Dy())-float64(o.State.Position.Y)-constants.PlayerHeight-offsetY)
	op.ColorScale.ScaleWithColor(color.White)
	text.DrawWithOptions(screen, t, f, op)

	if !o.State.IsDead() {
		// Draw hitpoints bar
		hitpointsBarWidth := float32(constants.NPCWidth)
		hitpointsBarHeight := float32(8)
		hitpointsBarYOffset := float32(12)
		hitpointsBarX := float32(o.State.Position.X)
		hitpointsBarY := float32(float64(screen.Bounds().Dy())-constants.NPCHeight) - float32(o.State.Position.Y) - hitpointsBarHeight - hitpointsBarYOffset
		hitpointsBarColor := color.RGBA{255, 0, 0, 255} // Red
		vector.DrawFilledRect(screen, hitpointsBarX, hitpointsBarY, hitpointsBarWidth, hitpointsBarHeight, hitpointsBarColor, false)

		// Draw hitpoints
		hitpointsWidth := float32(float64(hitpointsBarWidth) * (float64(o.State.Hitpoints) / float64(constants.PlayerHitpoints)))
		hitpointsHeight := float32(hitpointsBarHeight)
		hitpointsX := hitpointsBarX
		hitpointsY := hitpointsBarY
		hitpointsColor := color.RGBA{0, 255, 0, 255} // Green
		vector.DrawFilledRect(screen, hitpointsX, hitpointsY, hitpointsWidth, hitpointsHeight, hitpointsColor, false)
	}

	if o.debug {
		strokeWidth := float32(1)
		var playerColor color.RGBA
		if o.isLocalPlayer {
			// red if not on ground, blue if on ground
			playerColor = color.RGBA{255, 0, 0, 255} // Red
			if o.State.IsOnGround {
				playerColor = color.RGBA{0, 0, 255, 255} // Blue
			}
		} else {
			playerColor = color.RGBA{0, 255, 60, 255} // Green
			if o.State.IsOnGround {
				playerColor = color.RGBA{200, 0, 200, 255} // Purple
			}
		}
		vector.StrokeRect(screen, float32(o.State.Position.X), float32(float64(screen.Bounds().Dy())-constants.PlayerHeight)-float32(o.State.Position.Y), float32(constants.PlayerWidth), float32(constants.PlayerHeight), strokeWidth, playerColor, false)
	}
}

func (o *Player) InterpolateState(from *gametypes.PlayerState, to *gametypes.PlayerState, factor float64) {
	o.State.LastProcessedTimestamp = to.LastProcessedTimestamp
	// TODO: extend or shorten movement based on number of client updates processed to address jitter
	o.State.Position.X = from.Position.X + (to.Position.X-from.Position.X)*factor
	o.State.Position.Y = from.Position.Y + (to.Position.Y-from.Position.Y)*factor
	o.State.Velocity.X = to.Velocity.X
	o.State.Velocity.Y = to.Velocity.X
	o.State.IsOnGround = to.IsOnGround
	o.State.IsAttacking = to.IsAttacking
	o.State.Animation = to.Animation
	o.State.AnimationFlip = to.AnimationFlip
	o.State.AnimationSequence = to.AnimationSequence
	o.State.Hitpoints = to.Hitpoints
	o.State.Object.Position.X = o.State.Position.X
	o.State.Object.Position.Y = o.State.Position.Y
}

func (o *Player) ExtrapolateState(from *gametypes.PlayerState, to *gametypes.PlayerState, factor float64) {
	o.State.LastProcessedTimestamp = to.LastProcessedTimestamp
	o.State.Position.X = to.Position.X + (to.Position.X-from.Position.X)*factor
	o.State.Position.Y = to.Position.Y + (to.Position.Y-from.Position.Y)*factor
	o.State.Velocity.X = to.Velocity.X
	o.State.Velocity.Y = to.Velocity.Y
	o.State.IsOnGround = to.IsOnGround
	o.State.IsAttacking = to.IsAttacking
	o.State.Animation = to.Animation
	o.State.AnimationFlip = to.AnimationFlip
	o.State.AnimationSequence = to.AnimationSequence
	o.State.Hitpoints = to.Hitpoints
	o.State.Object.Position.X = o.State.Position.X
	o.State.Object.Position.Y = o.State.Position.Y
}

// ReconcileState reconciles the player state with the server state
// by going back through the past client states and checking if it
// the client state for that timestamp matches the server state.
// If it doesn't match, the server state is applied and all of the
// past updates that are after the last processed timestamp are replayed.
func (o *Player) ReconcileState(state *gametypes.PlayerState) error {
	if state.LastProcessedTimestamp == 0 {
		// initial state received from the server, nothing to reconcile
		return nil
	}

	// update pieces of the state that are not predicted (e.g. hitpoints)
	o.State.Hitpoints = state.Hitpoints

	foundPreviousState := false
	for i := len(o.previousStates) - 1; i >= 0; i-- {
		ps := o.previousStates[i]
		if ps.Timestamp == state.LastProcessedTimestamp {
			foundPreviousState = true
			if ps.State.NeedsReconciliation(state) {
				log.Warn("Reconciling player state at timestamp %d for %s", state.LastProcessedTimestamp, o.ID)
				log.Warn("Client state: %v", ps.State)
				log.Warn("Server state: %v", state)
				// apply the server state
				o.State.Position.X = state.Position.X
				o.State.Position.Y = state.Position.Y
				o.State.Velocity.X = state.Velocity.X
				o.State.Velocity.Y = state.Velocity.Y
				o.State.IsOnGround = state.IsOnGround
				o.State.IsAttacking = state.IsAttacking
				o.State.Animation = state.Animation
				o.State.AnimationFlip = state.AnimationFlip
				o.State.AnimationSequence = state.AnimationSequence
				o.State.Object.Position.X = state.Position.X
				o.State.Object.Position.Y = state.Position.Y

				// replay all of the past updates that are after the reconciled state
				for j := 0; j < len(o.pastUpdates); j++ {
					if o.pastUpdates[j].Timestamp > state.LastProcessedTimestamp {
						o.State.ApplyInput(o.pastUpdates[j])
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
