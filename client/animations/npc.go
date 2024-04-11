package animations

import (
	"bytes"
	"fmt"
	"image"

	"github.com/cbodonnell/flywheel/client/spritesheets"
	"github.com/hajimehoshi/ebiten/v2"
)

var (
	npcSpritesheet image.Image
)

func init() {
	var err error
	npcSpritesheet, _, err = image.Decode(bytes.NewReader(spritesheets.SkeletonIdle))
	if err != nil {
		panic(fmt.Sprintf("failed to decode image: %v", err))
	}
}

func NewNPCIdleAnimation() *Animation {
	return NewAnimation(NewAnimationOptions{
		Image:       ebiten.NewImageFromImage(npcSpritesheet),
		FrameOX:     0,
		FrameOY:     0,
		FrameWidth:  128,
		FrameHeight: 128,
		FrameCount:  7,
		FrameSpeed:  7,
		ScaleX:      0.5,
		ScaleY:      0.5,
		ShiftX:      -32,
		ShiftY:      0,
	})
}
