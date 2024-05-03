package scenes

import "github.com/cbodonnell/flywheel/client/objects"

type ErrorScene struct {
	*BaseScene
}

var _ Scene = &ErrorScene{}

func NewErrorScene(msg string) (Scene, error) {
	return &ErrorScene{
		BaseScene: NewBaseScene(objects.NewTextOverlayObject("overlay-error", msg)),
	}, nil
}
