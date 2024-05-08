package objects

import (
	"image/color"
	"strings"

	"github.com/cbodonnell/flywheel/client/animations"
	"github.com/cbodonnell/flywheel/client/fonts"
	"github.com/cbodonnell/flywheel/pkg/game/constants"
	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/solarlune/resolv"
	"golang.org/x/image/font"
)

type NPC struct {
	*BaseObject

	ID    string
	debug bool
	// TODO: make this private with a getter and setter
	State *gametypes.NPCState

	animations                 map[gametypes.NPCAnimation]*animations.Animation
	lastDrawnAnimationSequence uint8
}

func NewNPC(id string, state *gametypes.NPCState) (*NPC, error) {
	state.Object = resolv.NewObject(state.Position.X, state.Position.Y, constants.NPCWidth, constants.NPCHeight, gametypes.CollisionSpaceTagNPC)

	baseObjectOpts := &NewBaseObjectOpts{
		ZIndex: 10,
	}

	return &NPC{
		BaseObject: NewBaseObject(id, baseObjectOpts),
		ID:         id,
		// debug: true,
		State: state,
		animations: map[gametypes.NPCAnimation]*animations.Animation{
			gametypes.NPCAnimationIdle:    animations.NewNPCIdleAnimation(),
			gametypes.NPCAnimationWalk:    animations.NewNPCWalkAnimation(),
			gametypes.NPCAnimationDead:    animations.NewNPCDeadAnimation(),
			gametypes.NPCAnimationAttack1: animations.NewNPCAttack1Animation(),
			gametypes.NPCAnimationAttack2: animations.NewNPCAttack2Animation(),
			gametypes.NPCAnimationAttack3: animations.NewNPCAttack3Animation(),
		},
	}, nil
}

func (o *NPC) Update() error {
	o.animations[o.State.Animation].Update()
	return nil
}

func (o *NPC) Draw(screen *ebiten.Image) {
	if o.State.AnimationSequence != o.lastDrawnAnimationSequence {
		o.animations[o.State.Animation].Reset()
	}
	o.animations[o.State.Animation].Draw(screen, o.State.Position.X, o.State.Position.Y, o.State.AnimationFlip)
	o.lastDrawnAnimationSequence = o.State.AnimationSequence

	if !o.State.IsDead() {
		// Draw Name
		t := strings.ToUpper("Skeleton")
		f := fonts.TTFSmallFont
		bounds, _ := font.BoundString(f, t)
		op := &ebiten.DrawImageOptions{}
		offsetY := float64(24)
		op.GeoM.Translate(float64(o.State.Position.X)+constants.PlayerWidth/2-float64(bounds.Max.X>>6)/2, float64(screen.Bounds().Dy())-float64(o.State.Position.Y)-constants.PlayerHeight-offsetY)
		op.ColorScale.ScaleWithColor(color.White)
		text.DrawWithOptions(screen, t, f, op)

		// Draw hitpoints bar
		hitpointsBarWidth := float32(constants.NPCWidth)
		hitpointsBarHeight := float32(8)
		hitpointsBarYOffset := float32(12)
		hitpointsBarX := float32(o.State.Position.X)
		hitpointsBarY := float32(float64(screen.Bounds().Dy())-constants.NPCHeight) - float32(o.State.Position.Y) - hitpointsBarHeight - hitpointsBarYOffset
		hitpointsBarColor := color.RGBA{255, 0, 0, 255} // Red
		vector.DrawFilledRect(screen, hitpointsBarX, hitpointsBarY, hitpointsBarWidth, hitpointsBarHeight, hitpointsBarColor, false)

		// Draw hitpoints
		hitpointsWidth := float32(float64(hitpointsBarWidth) * (float64(o.State.Hitpoints) / float64(constants.NPCHitpoints)))
		hitpointsHeight := float32(hitpointsBarHeight)
		hitpointsX := hitpointsBarX
		hitpointsY := hitpointsBarY
		hitpointsColor := color.RGBA{0, 255, 0, 255} // Green
		vector.DrawFilledRect(screen, hitpointsX, hitpointsY, hitpointsWidth, hitpointsHeight, hitpointsColor, false)
	}

	if o.debug {
		strokeWidth := float32(1)
		npcColor := color.RGBA{0, 255, 60, 255} // Green
		if o.State.IsOnGround {
			npcColor = color.RGBA{200, 0, 200, 255} // Purple
		}
		vector.StrokeRect(screen, float32(o.State.Position.X), float32(float64(screen.Bounds().Dy())-constants.NPCHeight)-float32(o.State.Position.Y), float32(constants.NPCHeight), float32(constants.NPCWidth), strokeWidth, npcColor, false)
	}
}

func (o *NPC) InterpolateState(from *gametypes.NPCState, to *gametypes.NPCState, factor float64) {
	if from.IsDead() && !to.IsDead() {
		o.State.Position.X = to.Position.X
		o.State.Position.Y = to.Position.Y
	} else {
		o.State.Position.X = from.Position.X + (to.Position.X-from.Position.X)*factor
		o.State.Position.Y = from.Position.Y + (to.Position.Y-from.Position.Y)*factor
	}
	o.State.Velocity.X = to.Velocity.X
	o.State.Velocity.Y = to.Velocity.X
	o.State.IsOnGround = to.IsOnGround
	o.State.Animation = to.Animation
	o.State.AnimationFlip = to.AnimationFlip
	o.State.AnimationSequence = to.AnimationSequence
	o.State.Hitpoints = to.Hitpoints
	o.State.Object.Position.X = o.State.Position.X
	o.State.Object.Position.Y = o.State.Position.Y
}

func (o *NPC) ExtrapolateState(from *gametypes.NPCState, to *gametypes.NPCState, factor float64) {
	o.State.Position.X = to.Position.X + (to.Position.X-from.Position.X)*factor
	o.State.Position.Y = to.Position.Y + (to.Position.Y-from.Position.Y)*factor
	o.State.Velocity.X = to.Velocity.X
	o.State.Velocity.Y = to.Velocity.Y
	o.State.IsOnGround = to.IsOnGround
	o.State.Animation = to.Animation
	o.State.AnimationFlip = to.AnimationFlip
	o.State.AnimationSequence = to.AnimationSequence
	o.State.Hitpoints = to.Hitpoints
	o.State.Object.Position.X = o.State.Position.X
	o.State.Object.Position.Y = o.State.Position.Y
}
