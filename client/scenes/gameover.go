package scenes

import "github.com/cbodonnell/flywheel/client/objects"

type GameOverScene struct {
	*BaseScene
}

var _ Scene = &GameOverScene{}

func NewGameOverScene() (Scene, error) {
	return &GameOverScene{
		BaseScene: NewBaseScene(objects.NewTextOverlayObject("overlay-gameover", "Game Over!")),
	}, nil
}
