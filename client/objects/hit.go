package objects

import (
	"image/color"
	"strings"

	"github.com/cbodonnell/flywheel/client/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

type TextEffect struct {
	*BaseObject

	ID     string
	text   string
	x      float64
	y      float64
	color  color.Color
	scroll bool
}

type NewTextEffectOptions struct {
	Text   string
	X      float64
	Y      float64
	Color  color.Color
	Scroll bool
}

func NewTextEffect(id string, opts NewTextEffectOptions) *TextEffect {
	clr := opts.Color
	if clr == nil {
		clr = color.White
	}

	return &TextEffect{
		BaseObject: NewBaseObject(id),
		ID:         id,
		text:       opts.Text,
		x:          opts.X,
		y:          opts.Y,
		color:      opts.Color,
		scroll:     opts.Scroll,
	}
}

func (o *TextEffect) Update() error {
	if o.scroll {
		factor := float64(ebiten.TPS()) / 60
		o.y += 1 * factor
	}
	return nil
}

func (o *TextEffect) Draw(screen *ebiten.Image) {
	t := strings.ToUpper(o.text)
	f := fonts.TTFSmallFont
	bounds, _ := font.BoundString(f, t)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(o.x-float64(bounds.Max.X>>6)/2, float64(screen.Bounds().Dy())-o.y)
	op.ColorScale.ScaleWithColor(o.color)
	text.DrawWithOptions(screen, t, f, op)
}
