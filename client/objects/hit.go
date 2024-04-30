package objects

import (
	"image/color"
	"strings"

	"github.com/cbodonnell/flywheel/client/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

type Hit struct {
	*BaseObject

	ID   string
	text string
	x    float64
	y    float64
}

func NewHit(id string, text string, x, y float64) *Hit {
	return &Hit{
		BaseObject: NewBaseObject(id),
		ID:         id,
		text:       text,
		x:          x,
		y:          y,
	}
}

func (p *Hit) Update() error {
	factor := float64(ebiten.TPS()) / 60
	p.y += 1 * factor
	return nil
}

func (p *Hit) Draw(screen *ebiten.Image) {
	t := strings.ToUpper(p.text)
	f := fonts.TTFSmallFont
	bounds, _ := font.BoundString(f, t)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.x-float64(bounds.Max.X>>6)/2, float64(screen.Bounds().Dy())-p.y)
	op.ColorScale.ScaleWithColor(color.RGBA{255, 0, 0, 255}) // red
	text.DrawWithOptions(screen, t, f, op)
}
