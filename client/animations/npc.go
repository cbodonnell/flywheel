package animations

import (
	"bytes"
	"fmt"
	"image"

	"github.com/cbodonnell/flywheel/client/spritesheets"
	"github.com/hajimehoshi/ebiten/v2"
)

var (
	npcIdleSpritesheet    image.Image
	npcWalkSpritesheet    image.Image
	npcDeadSpritesheet    image.Image
	npcAttack1Spritesheet image.Image
)

func init() {
	var err error
	npcIdleSpritesheet, _, err = image.Decode(bytes.NewReader(spritesheets.SkeletonIdle))
	if err != nil {
		panic(fmt.Sprintf("failed to decode image: %v", err))
	}

	npcWalkSpritesheet, _, err = image.Decode(bytes.NewReader(spritesheets.SkeletonWalk))
	if err != nil {
		panic(fmt.Sprintf("failed to decode image: %v", err))
	}

	npcDeadSpritesheet, _, err = image.Decode(bytes.NewReader(spritesheets.SkeletonDead))
	if err != nil {
		panic(fmt.Sprintf("failed to decode image: %v", err))
	}

	npcAttack1Spritesheet, _, err = image.Decode(bytes.NewReader(spritesheets.SkeletonAttack1))
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

func NewNPCWalkAnimation() *Animation {
	return NewAnimation(NewAnimationOptions{
		Image:       ebiten.NewImageFromImage(npcWalkSpritesheet),
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

func NewNPCAttack1Animation() *Animation {
	return NewAnimation(NewAnimationOptions{
		Image:       ebiten.NewImageFromImage(npcAttack1Spritesheet),
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
		IsLooping:   false,
	})
}
