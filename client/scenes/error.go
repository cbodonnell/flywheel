package scenes

import "github.com/cbodonnell/flywheel/client/objects"

type ErrorScene struct {
	BaseScene
}

func NewErrorScene(msg string) (Scene, error) {
	return &MenuScene{
		BaseScene{
			Root: objects.NewTextOverlayObject("overlay-error", msg),
		},
	}, nil
}
