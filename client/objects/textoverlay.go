package objects

import (
	"image/color"
	"strings"

	"github.com/cbodonnell/flywheel/client/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

type TextOverlayObject struct {
	*BaseObject

	text string
}

func NewTextOverlayObject(id string, text string) GameObject {
	return &TextOverlayObject{
		BaseObject: NewBaseObject(id, nil),
		text:       text,
	}
}

func (o *TextOverlayObject) Draw(screen *ebiten.Image) {
	t := strings.ToUpper(o.text)
	f := fonts.TTFLargeFont
	bounds, _ := font.BoundString(f, t)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(screen.Bounds().Dx())/2-float64(bounds.Max.X>>6)/2, float64(screen.Bounds().Dy())/2-float64(bounds.Max.Y>>6)/2)
	op.ColorScale.ScaleWithColor(color.White)
	text.DrawWithOptions(screen, t, f, op)
}
