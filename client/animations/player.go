package animations

import (
	"bytes"
	"fmt"
	"image"

	"github.com/cbodonnell/flywheel/client/spritesheets"
	"github.com/hajimehoshi/ebiten/v2"
)

var (
	swordsmanIdleSpritesheet   image.Image
	swodsmanRunSpritesheet     image.Image
	swordsmanJumpSpritesheet   image.Image
	swordsmanAttackSpritesheet image.Image
)

func init() {
	var err error
	swordsmanIdleSpritesheet, _, err = image.Decode(bytes.NewReader(spritesheets.PlayerSwordsmanIdle))
	if err != nil {
		panic(fmt.Sprintf("failed to decode image: %v", err))
	}

	swodsmanRunSpritesheet, _, err = image.Decode(bytes.NewReader(spritesheets.PlayerSwordsmanRun))
	if err != nil {
		panic(fmt.Sprintf("failed to decode image: %v", err))
	}

	swordsmanJumpSpritesheet, _, err = image.Decode(bytes.NewReader(spritesheets.PlayerSwordsmanJump))
	if err != nil {
		panic(fmt.Sprintf("failed to decode image: %v", err))
	}

	swordsmanAttackSpritesheet, _, err = image.Decode(bytes.NewReader(spritesheets.PlayerSwordsmanAttack3))
	if err != nil {
		panic(fmt.Sprintf("failed to decode image: %v", err))
	}
}

func NewPlayerIdleAnimation() *Animation {
	return NewAnimation(NewAnimationOptions{
		Image:       ebiten.NewImageFromImage(swordsmanIdleSpritesheet),
		FrameOX:     0,
		FrameOY:     0,
		FrameWidth:  128,
		FrameHeight: 128,
		FrameCount:  8,
		FrameSpeed:  8,
		ScaleX:      1.0,
		ScaleY:      1.0,
		ShiftX:      -32,
		ShiftY:      0,
		IsLooping:   true,
	})
}

func NewPlayerRunAnimation() *Animation {
	return NewAnimation(NewAnimationOptions{
		Image:       ebiten.NewImageFromImage(swodsmanRunSpritesheet),
		FrameOX:     0,
		FrameOY:     0,
		FrameWidth:  128,
		FrameHeight: 128,
		FrameCount:  8,
		FrameSpeed:  4,
		ScaleX:      1.0,
		ScaleY:      1.0,
		ShiftX:      -32,
		ShiftY:      0,
		IsLooping:   true,
	})
}

func NewPlayerJumpAnimation() *Animation {
	return NewAnimation(NewAnimationOptions{
		Image:       ebiten.NewImageFromImage(swordsmanJumpSpritesheet),
		FrameOX:     0,
		FrameOY:     0,
		FrameWidth:  128,
		FrameHeight: 128,
		FrameCount:  4,
		FrameSpeed:  1,
		ScaleX:      1.0,
		ScaleY:      1.0,
		ShiftX:      -32,
		ShiftY:      0,
		IsLooping:   false,
	})
}

func NewPlayerFallAnimation() *Animation {
	return NewAnimation(NewAnimationOptions{
		Image:       ebiten.NewImageFromImage(swordsmanJumpSpritesheet),
		FrameOX:     128 * 3, // 4th frame
		FrameOY:     0,
		FrameWidth:  128,
		FrameHeight: 128,
		FrameCount:  4,
		FrameSpeed:  1,
		ScaleX:      1.0,
		ScaleY:      1.0,
		ShiftX:      -32,
		ShiftY:      0,
		IsLooping:   false,
	})
}

func NewPlayerAttackAnimation() *Animation {
	return NewAnimation(NewAnimationOptions{
		Image:       ebiten.NewImageFromImage(swordsmanAttackSpritesheet),
		FrameOX:     0,
		FrameOY:     0,
		FrameWidth:  128,
		FrameHeight: 128,
		FrameCount:  4,
		FrameSpeed:  4,
		ScaleX:      1.0,
		ScaleY:      1.0,
		ShiftX:      -32,
		ShiftY:      0,
		IsLooping:   false,
	})
}
