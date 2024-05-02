package scenes

import "github.com/cbodonnell/flywheel/client/objects"

type ErrorScene struct {
	BaseScene
}

func NewErrorScene(msg string) (Scene, error) {
	return &ErrorScene{
		BaseScene{
			root: objects.NewTextOverlayObject("overlay-error", msg),
		},
	}, nil
}
