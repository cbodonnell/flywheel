package objects

import (
	"image/color"

	"github.com/cbodonnell/flywheel/client/animations"
	"github.com/cbodonnell/flywheel/pkg/game/constants"
	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/solarlune/resolv"
)

type NPC struct {
	BaseObject

	ID    string
	debug bool
	// TODO: make this private with a getter and setter
	State *gametypes.NPCState

	animations map[gametypes.NPCAnimation]*animations.Animation
}

func NewNPC(id string, state *gametypes.NPCState) (*NPC, error) {
	state.Object = resolv.NewObject(state.Position.X, state.Position.Y, constants.NPCWidth, constants.NPCHeight, gametypes.CollisionSpaceTagNPC)

	return &NPC{
		BaseObject: BaseObject{
			Children: make(map[string]GameObject),
		},
		ID:    id,
		debug: false,
		State: state,
		animations: map[gametypes.NPCAnimation]*animations.Animation{
			gametypes.NPCAnimationIdle: animations.NewNPCIdleAnimation(),
			// gametypes.NPCAnimationRun:  animations.NewNPCRunAnimation(),
			// gametypes.NPCAnimationJump: animations.NewNPCJumpAnimation(),
			// gametypes.NPCAnimationFall: animations.NewNPCFallAnimation(),
		},
	}, nil
}

func (p *NPC) Update() error {
	p.animations[p.State.Animation].Update()
	return nil
}

func (p *NPC) Draw(screen *ebiten.Image) {
	npcObject := p.State.Object

	if p.debug {
		strokeWidth := float32(1)
		npcColor := color.RGBA{0, 255, 60, 255} // Green
		if p.State.IsOnGround {
			npcColor = color.RGBA{200, 0, 200, 255} // Purple
		}
		vector.StrokeRect(screen, float32(npcObject.Position.X), float32(float64(screen.Bounds().Dy())-npcObject.Size.Y)-float32(npcObject.Position.Y), float32(npcObject.Size.X), float32(npcObject.Size.Y), strokeWidth, npcColor, false)
	}

	frameWidth, frameHeight := p.animations[p.State.Animation].Size()
	scaleX, scaleY := 1.0, 1.0
	translateX := npcObject.Position.X
	translateY := float64(screen.Bounds().Dy()-frameHeight) - npcObject.Position.Y
	if p.State.AnimationFlip {
		scaleX = -1.0
		translateX = (-1.0 * translateX) - float64(frameWidth)
	}

	op := p.animations[p.State.Animation].DefaultOptions()
	op.GeoM.Translate(translateX, translateY)
	op.GeoM.Scale(scaleX, scaleY)
	screen.DrawImage(p.animations[p.State.Animation].CurrentImage(), op)
}

func (p *NPC) InterpolateState(from *gametypes.NPCState, to *gametypes.NPCState, factor float64) {
	p.State.Position.X = from.Position.X + (to.Position.X-from.Position.X)*factor
	p.State.Position.Y = from.Position.Y + (to.Position.Y-from.Position.Y)*factor
	p.State.Velocity.X = to.Velocity.X
	p.State.Velocity.Y = to.Velocity.X
	p.State.IsOnGround = to.IsOnGround
	p.State.Animation = to.Animation
	p.State.AnimationFlip = to.AnimationFlip
	p.State.Object.Position.X = p.State.Position.X
	p.State.Object.Position.Y = p.State.Position.Y
}

func (p *NPC) ExtrapolateState(from *gametypes.NPCState, to *gametypes.NPCState, factor float64) {
	p.State.Position.X = to.Position.X + (to.Position.X-from.Position.X)*factor
	p.State.Position.Y = to.Position.Y + (to.Position.Y-from.Position.Y)*factor
	p.State.Velocity.X = to.Velocity.X
	p.State.Velocity.Y = to.Velocity.Y
	p.State.IsOnGround = to.IsOnGround
	p.State.Animation = to.Animation
	p.State.AnimationFlip = to.AnimationFlip
	p.State.Object.Position.X = p.State.Position.X
	p.State.Object.Position.Y = p.State.Position.Y
}
