package animations

import (
	"bytes"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
)

var (
	npcSpritesheet image.Image
)

func init() {
	var err error
	npcSpritesheet, _, err = image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		panic(fmt.Sprintf("failed to decode image: %v", err))
	}
}

func NewNPCIdleAnimation() *Animation {
	return NewAnimation(NewAnimationOptions{
		Image:       ebiten.NewImageFromImage(npcSpritesheet),
		FrameOX:     0,
		FrameOY:     0,
		FrameWidth:  32,
		FrameHeight: 32,
		FrameCount:  5,
		FrameSpeed:  5,
	})
}
