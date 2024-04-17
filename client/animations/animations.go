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
	// scaleX is the horizontal scale factor.
	scaleX float64
	// scaleY is the vertical scale factor.
	scaleY float64
	// shiftX is the horizontal shift.
	shiftX float64
	// shiftY is the vertical shift.
	shiftY float64

	// isLooping is true if the animation should loop.
	isLooping bool

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
	ScaleX      float64
	ScaleY      float64
	ShiftX      float64
	ShiftY      float64
	IsLooping   bool
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
		scaleX:      opts.ScaleX,
		scaleY:      opts.ScaleY,
		shiftX:      opts.ShiftX,
		shiftY:      opts.ShiftY,
		isLooping:   opts.IsLooping,
	}
}

func (a *Animation) Update() {
	if a.updateCount%a.frameSpeed == 0 {
		a.frameIndex++
		if a.frameIndex >= a.frameCount {
			if a.isLooping {
				a.frameIndex = 0
			} else {
				a.frameIndex = a.frameCount - 1
			}
		}
	}
	a.updateCount++
}

func (a *Animation) IsFinished() bool {
	return a.frameIndex == a.frameCount-1
}

func (a *Animation) Reset() {
	a.updateCount = 0
	a.frameIndex = 0
}

func (a *Animation) Draw(screen *ebiten.Image, positionX float64, positionY float64, flip bool) {
	frameWidth, frameHeight := a.Size()
	scaleX, scaleY := a.Scale()
	shiftX, shiftY := a.Shift()
	translateX := positionX + (shiftX * scaleX)
	translateY := float64(screen.Bounds().Dy()) - (scaleY * float64(frameHeight)) - positionY - (shiftY * scaleY)
	if flip {
		scaleX = -scaleX
		translateX = translateX + (-scaleX * float64(frameWidth))
	}

	op := a.DefaultOptions()
	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(translateX, translateY)
	screen.DrawImage(a.CurrentImage(), op)
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

func (a *Animation) Size() (w int, h int) {
	return a.frameWidth, a.frameHeight
}

func (a *Animation) Scale() (x float64, y float64) {
	return a.scaleX, a.scaleY
}

func (a *Animation) Shift() (x float64, y float64) {
	return a.shiftX, a.shiftY
}
