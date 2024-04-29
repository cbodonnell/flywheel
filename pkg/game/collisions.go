package game

import (
	"github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/solarlune/resolv"
)

// TODO: make this dynamic
func NewCollisionSpace() *resolv.Space {
	spaceWidth, spaceHeight := 640, 480
	cellWidth, cellHeight := 16, 16
	space := resolv.NewSpace(spaceWidth, spaceHeight, cellWidth, cellHeight)
	space.Add(
		resolv.NewObject(0, 0, 640, 16, types.CollisionSpaceTagLevel),
		resolv.NewObject(0, 480-16, 640, 16, types.CollisionSpaceTagLevel),
		resolv.NewObject(0, 16, 16, 480-32, types.CollisionSpaceTagLevel),
		resolv.NewObject(640-16, 16, 16, 480-32, types.CollisionSpaceTagLevel),
		// add a platform
		resolv.NewObject(320-64, 96, 128, 16, types.CollisionSpaceTagLevel),
	)
	return space
}
