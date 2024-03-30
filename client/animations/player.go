package animations

import (
	"bytes"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
)

var (
	PlayerIdleAnimation *Animation
	PlayerRunAnimation  *Animation
	PlayerJumpAnimation *Animation
)

func init() {
	var err error
	PlayerIdleAnimation, err = NewPlayerIdleAnimation()
	if err != nil {
		panic(fmt.Sprintf("failed to create player idle animation: %v", err))
	}
	PlayerRunAnimation, err = NewPlayerRunAnimation()
	if err != nil {
		panic(fmt.Sprintf("failed to create player run animation: %v", err))
	}
	PlayerJumpAnimation, err = NewPlayerJumpAnimation()
	if err != nil {
		panic(fmt.Sprintf("failed to create player jump animation: %v", err))
	}
}

func NewPlayerIdleAnimation() (*Animation, error) {
	img, _, err := image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	return NewAnimation(NewAnimationOptions{
		Image:       ebiten.NewImageFromImage(img),
		FrameOX:     0,
		FrameOY:     0,
		FrameWidth:  32,
		FrameHeight: 32,
		FrameCount:  5,
		FrameSpeed:  5,
	}), nil
}

func NewPlayerRunAnimation() (*Animation, error) {
	img, _, err := image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	return NewAnimation(NewAnimationOptions{
		Image:       ebiten.NewImageFromImage(img),
		FrameOX:     0,
		FrameOY:     32,
		FrameWidth:  32,
		FrameHeight: 32,
		FrameCount:  8,
		FrameSpeed:  5,
	}), nil
}

func NewPlayerJumpAnimation() (*Animation, error) {
	img, _, err := image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	return NewAnimation(NewAnimationOptions{
		Image:       ebiten.NewImageFromImage(img),
		FrameOX:     0,
		FrameOY:     64,
		FrameWidth:  32,
		FrameHeight: 32,
		FrameCount:  4,
		FrameSpeed:  5,
	}), nil
}
