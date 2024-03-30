package animations

import (
	"bytes"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
)

var PlayerIdleAnimation *Animation
var PlayerRunAnimation *Animation
var PlayerJumpAnimation *Animation

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

	frameOX := 0
	frameOY := 0
	frameWidth := 32
	frameHeight := 32
	frameCount := 5

	return NewAnimation(ebiten.NewImageFromImage(img), frameOX, frameOY, frameWidth, frameHeight, frameCount), nil
}

func NewPlayerRunAnimation() (*Animation, error) {
	img, _, err := image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	frameOX := 0
	frameOY := 32
	frameWidth := 32
	frameHeight := 32
	frameCount := 8

	return NewAnimation(ebiten.NewImageFromImage(img), frameOX, frameOY, frameWidth, frameHeight, frameCount), nil
}

func NewPlayerJumpAnimation() (*Animation, error) {
	img, _, err := image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	frameOX := 0
	frameOY := 64
	frameWidth := 32
	frameHeight := 32
	frameCount := 4

	return NewAnimation(ebiten.NewImageFromImage(img), frameOX, frameOY, frameWidth, frameHeight, frameCount), nil
}
