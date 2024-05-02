package scenes

import "github.com/cbodonnell/flywheel/client/objects"

type GameOverScene struct {
	BaseScene
}

func NewGameOverScene() (Scene, error) {
	return &GameOverScene{
		BaseScene{
			root: objects.NewTextOverlayObject("overlay-gameover", "Game Over!"),
		},
	}, nil
}
