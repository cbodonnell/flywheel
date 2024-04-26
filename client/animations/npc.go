package animations

import (
	"bytes"
	"fmt"
	"image"

	"github.com/cbodonnell/flywheel/client/spritesheets"
	"github.com/hajimehoshi/ebiten/v2"
)

var (
	npcIdleSpritesheet image.Image
	npcDeadSpritesheet image.Image
)

func init() {
	var err error
	npcIdleSpritesheet, _, err = image.Decode(bytes.NewReader(spritesheets.SkeletonIdle))
	if err != nil {
		panic(fmt.Sprintf("failed to decode image: %v", err))
	}

	npcDeadSpritesheet, _, err = image.Decode(bytes.NewReader(spritesheets.SkeletonDead))
	if err != nil {
		panic(fmt.Sprintf("failed to decode image: %v", err))
	}
}

func NewNPCIdleAnimation() *Animation {
	return NewAnimation(NewAnimationOptions{
		Image:       ebiten.NewImageFromImage(npcIdleSpritesheet),
		FrameOX:     0,
		FrameOY:     0,
		FrameWidth:  128,
		FrameHeight: 128,
		FrameCount:  7,
		FrameSpeed:  7,
		ScaleX:      1.0,
		ScaleY:      1.0,
		ShiftX:      -32,
		ShiftY:      0,
		IsLooping:   true,
	})
}

func NewNPCDeadAnimation() *Animation {
	return NewAnimation(NewAnimationOptions{
		Image:       ebiten.NewImageFromImage(npcDeadSpritesheet),
		FrameOX:     0,
		FrameOY:     0,
		FrameWidth:  128,
		FrameHeight: 128,
		FrameCount:  3,
		FrameSpeed:  3,
		ScaleX:      1.0,
		ScaleY:      1.0,
		ShiftX:      -32,
		ShiftY:      0,
		IsLooping:   false,
	})
}
