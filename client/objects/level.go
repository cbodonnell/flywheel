package objects

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type LevelObject struct {
	*BaseObject

	x, y float32
	w, h float32
	clr  color.Color
}

type NewLevelObjectOptions struct {
	// X is the x-coordinate of the level object.
	X float32
	// Y is the y-coordinate of the level object.
	Y float32
	// W is the width of the level object.
	W float32
	// H is the height of the level object.
	H float32
	// Color is the color of the level object.
	Color color.Color
	// ZIndex is the z-index of the level object.
	ZIndex int
}

func NewLevelObject(id string, opts NewLevelObjectOptions) *LevelObject {
	return &LevelObject{
		BaseObject: NewBaseObject(id, &NewBaseObjectOpts{
			ZIndex: opts.ZIndex,
		}),
		x:   opts.X,
		y:   opts.Y,
		w:   opts.W,
		h:   opts.H,
		clr: opts.Color,
	}
}

func (o *LevelObject) Draw(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, o.x, o.y, o.w, o.h, o.clr, false)
}
