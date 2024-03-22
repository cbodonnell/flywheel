package objects

import "github.com/hajimehoshi/ebiten/v2"

type Lifecycle interface {
	// Game flow methods
	Init() error
	Destroy() error
	Update() error
	Draw(screen *ebiten.Image)
}
