package animations

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type Animation struct {
	// image is the image containing the animation frames.
	image *ebiten.Image
	// frameOX is the x offset of the first frame in the animation.
	frameOX int
	// frameOY is the y offset of the first frame in the animation.
	frameOY int
	// frameWidth is the width of each frame in the animation.
	frameWidth int
	// frameHeight is the height of each frame in the animation.
	frameHeight int
	// frameCount is the number of frames in the animation.
	frameCount int
	// frameSpeed is the number of updates before the frame index is incremented.
	frameSpeed int

	// updateCount is the number of times the animation has been updated.
	updateCount int
	// frameIndex is the current frame index.
	frameIndex int
}

type NewAnimationOptions struct {
	Image       *ebiten.Image
	FrameOX     int
	FrameOY     int
	FrameWidth  int
	FrameHeight int
	FrameCount  int
	FrameSpeed  int
}

func NewAnimation(opts NewAnimationOptions) *Animation {
	return &Animation{
		image:       opts.Image,
		frameOX:     opts.FrameOX,
		frameOY:     opts.FrameOY,
		frameWidth:  opts.FrameWidth,
		frameHeight: opts.FrameHeight,
		frameCount:  opts.FrameCount,
		frameSpeed:  opts.FrameSpeed,
	}
}

func (a *Animation) Update() {
	a.updateCount++
	a.frameIndex = (a.updateCount / a.frameSpeed) % a.frameCount
}

func (a *Animation) Reset() {
	a.updateCount = 0
	a.frameIndex = 0
}

func (a *Animation) DefaultOptions() *ebiten.DrawImageOptions {
	return &ebiten.DrawImageOptions{
		Filter: ebiten.FilterNearest,
	}
}

func (a *Animation) CurrentImage() *ebiten.Image {
	sx, sy := a.frameOX+a.frameIndex*a.frameWidth, a.frameOY
	return a.image.SubImage(image.Rect(sx, sy, sx+a.frameWidth, sy+a.frameHeight)).(*ebiten.Image)
}

func (a *Animation) Size() (int, int) {
	return a.frameWidth, a.frameHeight
}
