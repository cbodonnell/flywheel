package scenes

import (
	"github.com/cbodonnell/flywheel/client/objects"
	"github.com/hajimehoshi/ebiten/v2"
)

type Scene interface {
	objects.Lifecycle

	// Scene specific methods
	GetRoot() objects.GameObject
}

type BaseScene struct {
	root objects.GameObject
}

func (s *BaseScene) GetRoot() objects.GameObject {
	return s.root
}

func (s *BaseScene) Init() error {
	return objects.InitTree(s.root)
}

func (s *BaseScene) Destroy() error {
	return objects.DestroyTree(s.root)
}

func (s *BaseScene) Update() error {
	return objects.UpdateTree(s.root)
}

func (s *BaseScene) Draw(screen *ebiten.Image) {
	objects.DrawTree(s.root, screen)
}
