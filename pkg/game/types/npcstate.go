package types

import "github.com/solarlune/resolv"

type NPCState struct {
	LastProcessedTimestamp int64    `json:"lastProcessedTimestamp"`
	Position               Position `json:"position"`
	Velocity               Velocity `json:"velocity"`
	// TODO: there's some redundancy here with the object reference
	Object        *resolv.Object  `json:"-"`
	IsOnGround    bool            `json:"isOnGround"`
	Animation     PlayerAnimation `json:"animation"`
	AnimationFlip bool            `json:"animationFlip"`
}
