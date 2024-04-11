package animations

import (
	"bytes"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
)

var (
	playerSpritesheet image.Image
)

func init() {
	var err error
	playerSpritesheet, _, err = image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		panic(fmt.Sprintf("failed to decode image: %v", err))
	}
}

func NewPlayerIdleAnimation() *Animation {
	return NewAnimation(NewAnimationOptions{
		Image:       ebiten.NewImageFromImage(playerSpritesheet),
		FrameOX:     0,
		FrameOY:     0,
		FrameWidth:  32,
		FrameHeight: 32,
		FrameCount:  5,
		FrameSpeed:  5,
		ScaleX:      1.0,
		ScaleY:      1.0,
		ShiftX:      0,
		ShiftY:      0,
	})
}

func NewPlayerRunAnimation() *Animation {
	return NewAnimation(NewAnimationOptions{
		Image:       ebiten.NewImageFromImage(playerSpritesheet),
		FrameOX:     0,
		FrameOY:     32,
		FrameWidth:  32,
		FrameHeight: 32,
		FrameCount:  8,
		FrameSpeed:  5,
		ScaleX:      1.0,
		ScaleY:      1.0,
		ShiftX:      0,
		ShiftY:      0,
	})
}

func NewPlayerJumpAnimation() *Animation {
	return NewAnimation(NewAnimationOptions{
		Image:       ebiten.NewImageFromImage(playerSpritesheet),
		FrameOX:     0,
		FrameOY:     64,
		FrameWidth:  32,
		FrameHeight: 32,
		FrameCount:  4,
		FrameSpeed:  5,
		ScaleX:      1.0,
		ScaleY:      1.0,
		ShiftX:      0,
		ShiftY:      0,
	})
}
