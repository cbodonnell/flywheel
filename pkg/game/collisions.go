package game

import (
	"github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/solarlune/resolv"
)

// TODO: make this dynamic
func NewCollisionSpace() *resolv.Space {
	space := resolv.NewSpace(640, 480, 16, 16)
	space.Add(
		resolv.NewObject(0, 0, 640, 16, types.CollisionSpaceTagLevel),
		resolv.NewObject(0, 480-16, 640, 16, types.CollisionSpaceTagLevel),
		resolv.NewObject(0, 16, 16, 480-32, types.CollisionSpaceTagLevel),
		resolv.NewObject(640-16, 16, 16, 480-32, types.CollisionSpaceTagLevel),
	)
	return space
}
