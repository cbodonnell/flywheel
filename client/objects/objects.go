package objects

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type GameObject interface {
	Update() error
	Draw(screen *ebiten.Image)
}
