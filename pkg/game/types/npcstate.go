package types

import (
	"github.com/cbodonnell/flywheel/pkg/game/constants"
	"github.com/solarlune/resolv"
)

type NPCState struct {
	Position Position `json:"position"`
	Velocity Velocity `json:"velocity"`
	// TODO: there's some redundancy here with the object reference
	Object        *resolv.Object  `json:"-"`
	IsOnGround    bool            `json:"isOnGround"`
	Animation     PlayerAnimation `json:"animation"`
	AnimationFlip bool            `json:"animationFlip"`
}

func NewNPCState() *NPCState {
	return &NPCState{
		Position: Position{
			X: constants.NPCStartingX,
			Y: constants.NPCStartingY,
		},
		Velocity: Velocity{
			X: 0,
			Y: 0,
		},
		Object: resolv.NewObject(constants.NPCStartingX, constants.NPCStartingY, constants.NPCWidth, constants.NPCHeight, CollisionSpaceTagPlayer),
	}
}
