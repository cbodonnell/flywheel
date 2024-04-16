package scenes

import "github.com/cbodonnell/flywheel/client/objects"

type GameOverScene struct {
	BaseScene
}

func NewGameOverScene() (Scene, error) {
	return &MenuScene{
		BaseScene{
			Root: objects.NewTextOverlayObject("overlay-gameover", "Game Over!"),
		},
	}, nil
}
