package objects

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// GameObject is the highest level interface for game related types.
type GameObject interface {
	Update() error
	Draw(screen *ebiten.Image)
}
