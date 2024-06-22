package game

import (
	"github.com/cbodonnell/flywheel/pkg/game/constants"
	"github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/solarlune/resolv"
)

// TODO: make this dynamic
func NewCollisionSpace() *resolv.Space {
	spaceWidth, spaceHeight := constants.SpaceWidth, constants.SpaceHeight
	cellWidth, cellHeight := constants.CellWidth, constants.CellHeight

	levelObjects := []*resolv.Object{
		// walls
		resolv.NewObject(0, 0, float64(spaceWidth), float64(cellHeight), types.CollisionSpaceTagLevel),
		resolv.NewObject(0, float64(spaceHeight-cellHeight), float64(spaceWidth), float64(cellHeight), types.CollisionSpaceTagLevel),
		resolv.NewObject(0, float64(cellHeight), float64(cellWidth), float64(spaceHeight-cellHeight*2), types.CollisionSpaceTagLevel),
		resolv.NewObject(float64(spaceWidth-cellWidth), float64(cellHeight), float64(cellWidth), float64(spaceHeight-cellHeight*2), types.CollisionSpaceTagLevel),
		// ladders
		resolv.NewObject(float64(spaceWidth*3/4-constants.CellWidth/2), float64(cellHeight*6), float64(cellWidth), float64(cellHeight*13), types.CollisionSpaceTagLadder),
		// platforms
		resolv.NewObject(float64(spaceWidth/2-cellWidth*4), float64(cellHeight*6), float64(cellWidth*8), float64(cellHeight), types.CollisionSpaceTagPlatform),
		resolv.NewObject(float64(spaceWidth/4-cellWidth*4), float64(cellHeight*12), float64(cellWidth*8), float64(cellHeight), types.CollisionSpaceTagPlatform),
		resolv.NewObject(float64(spaceWidth*3/4-cellWidth*4), float64(cellHeight*18), float64(cellWidth*8), float64(cellHeight), types.CollisionSpaceTagPlatform),
	}

	for _, obj := range levelObjects {
		obj.SetShape(resolv.NewRectangle(0, 0, obj.Size.X, obj.Size.Y))
	}

	space := resolv.NewSpace(spaceWidth, spaceHeight, cellWidth, cellHeight)
	space.Add(levelObjects...)
	return space
}
