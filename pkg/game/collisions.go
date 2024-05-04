package game

import (
	"github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/solarlune/resolv"
)

// TODO: make this dynamic
func NewCollisionSpace() *resolv.Space {
	spaceWidth, spaceHeight := 1280, 480
	cellWidth, cellHeight := 16, 16
	space := resolv.NewSpace(spaceWidth, spaceHeight, cellWidth, cellHeight)
	space.Add(
		resolv.NewObject(0, 0, float64(spaceWidth), float64(cellHeight), types.CollisionSpaceTagLevel),
		resolv.NewObject(0, float64(spaceHeight-cellHeight), float64(spaceWidth), float64(cellHeight), types.CollisionSpaceTagLevel),
		resolv.NewObject(0, float64(cellHeight), float64(cellWidth), float64(spaceHeight-cellHeight*2), types.CollisionSpaceTagLevel),
		resolv.NewObject(float64(spaceWidth-cellWidth), float64(cellHeight), float64(cellWidth), float64(spaceHeight-cellHeight*2), types.CollisionSpaceTagLevel),
		// add a platform
		resolv.NewObject(float64(spaceWidth/2-cellWidth*4), float64(cellHeight*6), float64(cellWidth*8), float64(cellHeight), types.CollisionSpaceTagLevel),
	)
	return space
}
