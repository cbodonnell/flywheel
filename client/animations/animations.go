package animations

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type Animation struct {
	image       *ebiten.Image
	frameOX     int
	frameOY     int
	frameWidth  int
	frameHeight int
	frameCount  int

	frameIndex int
}

func NewAnimation(image *ebiten.Image, frameOX, frameOY, frameWidth, frameHeight, frameCount int) *Animation {
	return &Animation{
		image:       image,
		frameOX:     frameOX,
		frameOY:     frameOY,
		frameWidth:  frameWidth,
		frameHeight: frameHeight,
		frameCount:  frameCount,
	}
}

func (a *Animation) Update() {
	a.frameIndex++
}

func (a *Animation) Reset() {
	a.frameIndex = 0
}

func (a *Animation) DefaultOptions() *ebiten.DrawImageOptions {
	return &ebiten.DrawImageOptions{
		Filter: ebiten.FilterNearest,
	}
}

func (a *Animation) CurrentImage() *ebiten.Image {
	i := (a.frameIndex / 5) % a.frameCount
	sx, sy := a.frameOX+i*a.frameWidth, a.frameOY
	return a.image.SubImage(image.Rect(sx, sy, sx+a.frameWidth, sy+a.frameHeight)).(*ebiten.Image)
}

func (a *Animation) Size() (int, int) {
	return a.frameWidth, a.frameHeight
}
