package objects

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type UIObject struct {
	*BaseObject
}

func NewUIObject(id string) GameObject {
	return &UIObject{
		BaseObject: NewBaseObject(id),
	}
}

func (o *UIObject) Draw(screen *ebiten.Image) {

}
