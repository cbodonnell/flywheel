package types

import "github.com/solarlune/resolv"

type PlayerState struct {
	LastProcessedTimestamp int64    `json:"lastProcessedTimestamp"`
	Position               Position `json:"position"`
	Velocity               Velocity `json:"velocity"`
	// TODO: there's some redundancy here with the object reference
	Object        *resolv.Object  `json:"-"`
	IsOnGround    bool            `json:"isOnGround"`
	Animation     PlayerAnimation `json:"animation"`
	AnimationFlip bool            `json:"animationFlip"`
}

type PlayerAnimation uint8

const (
	PlayerAnimationIdle PlayerAnimation = iota
	PlayerAnimationRun
	PlayerAnimationJump
	PlayerAnimationFall
)

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Velocity struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Equal returns true if the player state is equal to the other player state
func (p *PlayerState) Equal(other *PlayerState) bool {
	return p.Position.X == other.Position.X &&
		p.Position.Y == other.Position.Y &&
		p.Velocity.X == other.Velocity.X &&
		p.Velocity.Y == other.Velocity.Y &&
		p.IsOnGround == other.IsOnGround &&
		p.Animation == other.Animation &&
		p.AnimationFlip == other.AnimationFlip
}

// Copy returns a copy of the player state with an empty object reference
func (p *PlayerState) Copy() *PlayerState {
	return &PlayerState{
		LastProcessedTimestamp: p.LastProcessedTimestamp,
		Position:               p.Position,
		Velocity:               p.Velocity,
		IsOnGround:             p.IsOnGround,
		Animation:              p.Animation,
		AnimationFlip:          p.AnimationFlip,
	}
}
