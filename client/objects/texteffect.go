package objects

import (
	"fmt"
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
	ttl    int
}

type NewTextEffectOptions struct {
	// Text is the text to display.
	Text string
	// X is the x-coordinate of the text.
	X float64
	// Y is the y-coordinate of the text.
	Y float64
	// Color is the color of the text.
	Color color.Color
	// Scroll is a boolean value indicating whether the text should scroll.
	Scroll bool
	// TTL is the time to live in milliseconds.
	TTL int
	// ZIndex is the z-index of the text effect.
	ZIndex int
}

func NewTextEffect(id string, opts NewTextEffectOptions) *TextEffect {
	clr := opts.Color
	if clr == nil {
		clr = color.White
	}

	baseObjectOpts := &NewBaseObjectOpts{
		ZIndex: opts.ZIndex,
	}

	return &TextEffect{
		BaseObject: NewBaseObject(id, baseObjectOpts),
		ID:         id,
		text:       opts.Text,
		x:          opts.X,
		y:          opts.Y,
		color:      opts.Color,
		scroll:     opts.Scroll,
		ttl:        opts.TTL,
	}
}

func (o *TextEffect) Update() error {
	if o.scroll {
		factor := float64(ebiten.TPS()) / 60
		o.y += 1 * factor
	}
	if o.ttl > 0 {
		o.ttl -= 1000 / ebiten.TPS()
		if o.ttl <= 0 {
			if err := o.BaseObject.RemoveFromParent(); err != nil {
				return fmt.Errorf("failed to remove text effect from parent: %w", err)
			}
		}
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
