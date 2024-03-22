package objects

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Scene interface {
	Lifecycle

	// Scene specific methods
	GetRoot() GameObject
}

type BaseScene struct {
	Root GameObject
}

func (s *BaseScene) GetRoot() GameObject {
	return s.Root
}

func (s *BaseScene) Init() error {
	return InitTree(s.Root)
}

func (s *BaseScene) Destroy() error {
	return DestroyTree(s.Root)
}

func (s *BaseScene) Update() error {
	return UpdateTree(s.Root)
}

func (s *BaseScene) Draw(screen *ebiten.Image) {
	DrawTree(s.Root, screen)
}

type MenuScene struct {
	BaseScene
}

func NewMenuScene() (Scene, error) {
	return &MenuScene{
		BaseScene{
			Root: &BaseObject{},
		},
	}, nil
}

// TODO: implement menu scene methods

type GameScene struct {
	BaseScene
}

func NewGameScene() (Scene, error) {
	return &GameScene{
		BaseScene{
			Root: &BaseObject{},
		},
	}, nil
}

type GameOverScene struct {
	BaseScene
}

func NewGameOverScene() (Scene, error) {
	return &GameOverScene{
		BaseScene{
			Root: &BaseObject{},
		},
	}, nil
}
