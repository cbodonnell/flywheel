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

	animations map[gametypes.NPCAnimation]*animations.Animation
}

func NewNPC(id string, state *gametypes.NPCState) (*NPC, error) {
	state.Object = resolv.NewObject(state.Position.X, state.Position.Y, constants.NPCWidth, constants.NPCHeight, gametypes.CollisionSpaceTagNPC)

	return &NPC{
		BaseObject: NewBaseObject(id),
		ID:         id,
		// debug: true,
		State: state,
		animations: map[gametypes.NPCAnimation]*animations.Animation{
			gametypes.NPCAnimationIdle: animations.NewNPCIdleAnimation(),
			gametypes.NPCAnimationDead: animations.NewNPCDeadAnimation(),
		},
	}, nil
}

func (p *NPC) Update() error {
	p.animations[p.State.Animation].Update()
	return nil
}

func (p *NPC) Draw(screen *ebiten.Image) {
	p.animations[p.State.Animation].Draw(screen, p.State.Position.X, p.State.Position.Y, p.State.AnimationFlip)
	for a, anim := range p.animations {
		if a == p.State.Animation {
			continue
		}
		anim.Reset()
	}

	if !p.State.IsDead() {
		// Draw Name
		t := strings.ToUpper("Skeleton")
		f := fonts.TTFSmallFont
		bounds, _ := font.BoundString(f, t)
		op := &ebiten.DrawImageOptions{}
		offsetY := float64(24)
		op.GeoM.Translate(float64(p.State.Position.X)+constants.PlayerWidth/2-float64(bounds.Max.X>>6)/2, float64(screen.Bounds().Dy())-float64(p.State.Position.Y)-constants.PlayerHeight-offsetY)
		op.ColorScale.ScaleWithColor(color.White)
		text.DrawWithOptions(screen, t, f, op)

		// Draw hitpoints bar
		hitpointsBarWidth := float32(constants.NPCWidth)
		hitpointsBarHeight := float32(8)
		hitpointsBarYOffset := float32(12)
		hitpointsBarX := float32(p.State.Position.X)
		hitpointsBarY := float32(float64(screen.Bounds().Dy())-constants.NPCHeight) - float32(p.State.Position.Y) - hitpointsBarHeight - hitpointsBarYOffset
		hitpointsBarColor := color.RGBA{255, 0, 0, 255} // Red
		vector.DrawFilledRect(screen, hitpointsBarX, hitpointsBarY, hitpointsBarWidth, hitpointsBarHeight, hitpointsBarColor, false)

		// Draw hitpoints
		hitpointsWidth := float32(float64(hitpointsBarWidth) * (float64(p.State.Hitpoints) / float64(constants.NPCHitpoints)))
		hitpointsHeight := float32(hitpointsBarHeight)
		hitpointsX := hitpointsBarX
		hitpointsY := hitpointsBarY
		hitpointsColor := color.RGBA{0, 255, 0, 255} // Green
		vector.DrawFilledRect(screen, hitpointsX, hitpointsY, hitpointsWidth, hitpointsHeight, hitpointsColor, false)
	}

	if p.debug {
		strokeWidth := float32(1)
		npcColor := color.RGBA{0, 255, 60, 255} // Green
		if p.State.IsOnGround {
			npcColor = color.RGBA{200, 0, 200, 255} // Purple
		}
		vector.StrokeRect(screen, float32(p.State.Position.X), float32(float64(screen.Bounds().Dy())-constants.NPCHeight)-float32(p.State.Position.Y), float32(constants.NPCHeight), float32(constants.NPCWidth), strokeWidth, npcColor, false)
	}
}

func (p *NPC) InterpolateState(from *gametypes.NPCState, to *gametypes.NPCState, factor float64) {
	p.State.Position.X = from.Position.X + (to.Position.X-from.Position.X)*factor
	p.State.Position.Y = from.Position.Y + (to.Position.Y-from.Position.Y)*factor
	p.State.Velocity.X = to.Velocity.X
	p.State.Velocity.Y = to.Velocity.X
	p.State.IsOnGround = to.IsOnGround
	p.State.Animation = to.Animation
	p.State.AnimationFlip = to.AnimationFlip
	p.State.Hitpoints = to.Hitpoints
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
	p.State.Hitpoints = to.Hitpoints
	p.State.Object.Position.X = p.State.Position.X
	p.State.Object.Position.Y = p.State.Position.Y
}
