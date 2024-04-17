package scenes

import (
	"github.com/cbodonnell/flywheel/client/objects"
)

type MenuScene struct {
	BaseScene
}

func NewMenuScene() (Scene, error) {
	return &MenuScene{
		BaseScene{
			Root: objects.NewTextOverlayObject("overlay-menu", "Press to Start!"),
		},
	}, nil
}
