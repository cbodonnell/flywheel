package types

import (
	"github.com/cbodonnell/flywheel/pkg/game/constants"
	"github.com/cbodonnell/flywheel/pkg/kinematic"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/solarlune/resolv"
)

type PlayerState struct {
	LastProcessedTimestamp int64
	Position               kinematic.Vector
	Velocity               kinematic.Vector
	Object                 *resolv.Object
	IsOnGround             bool
	IsAttacking            bool
	AttackTimeLeft         float64
	IsAttackHitting        bool
	DidAttackHit           bool
	Animation              PlayerAnimation
	AnimationFlip          bool
}

type PlayerAnimation uint8

const (
	PlayerAnimationIdle PlayerAnimation = iota
	PlayerAnimationRun
	PlayerAnimationJump
	PlayerAnimationFall
	PlayerAnimationAttack
)

func NewPlayerState(positionX, positionY float64) *PlayerState {
	return &PlayerState{
		Position: kinematic.Vector{
			X: positionX,
			Y: positionY,
		},
		Velocity: kinematic.Vector{
			X: 0,
			Y: 0,
		},
		Object: resolv.NewObject(positionX, positionY, constants.PlayerWidth, constants.PlayerHeight, CollisionSpaceTagPlayer),
	}
}

// Equal returns true if the player state is equal to the other player state
func (p *PlayerState) Equal(other *PlayerState) bool {
	return p.Position.X == other.Position.X &&
		p.Position.Y == other.Position.Y &&
		p.Velocity.X == other.Velocity.X &&
		p.Velocity.Y == other.Velocity.Y &&
		p.IsOnGround == other.IsOnGround &&
		p.IsAttacking == other.IsAttacking &&
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
		IsAttacking:            p.IsAttacking,
		Animation:              p.Animation,
		AnimationFlip:          p.AnimationFlip,
	}
}

// ApplyInput updates the player's state based on the client's input.
// The player state is updated in place.
func (p *PlayerState) ApplyInput(clientPlayerUpdate *messages.ClientPlayerUpdate) {
	// Attack

	if p.AttackTimeLeft > 0 {
		p.AttackTimeLeft -= clientPlayerUpdate.DeltaTime
		if !p.DidAttackHit {
			if p.AttackTimeLeft <= constants.PlayerAttackDuration-constants.PlayerAttackChannelTime {
				// register the hit only once
				p.IsAttackHitting = true
				p.DidAttackHit = true
			}
		} else {
			p.IsAttackHitting = false
		}
	} else {
		p.IsAttacking = false
		p.IsAttackHitting = false
		p.DidAttackHit = false
	}

	if !p.IsAttacking && clientPlayerUpdate.InputAttack {
		p.IsAttacking = true
		p.AttackTimeLeft = constants.PlayerAttackDuration
	}

	// Movement

	// X-axis
	var dx, vx float64
	if !p.IsAttacking {
		dx = kinematic.Displacement(clientPlayerUpdate.InputX*constants.PlayerSpeed, clientPlayerUpdate.DeltaTime, 0)
		vx = kinematic.FinalVelocity(clientPlayerUpdate.InputX*constants.PlayerSpeed, clientPlayerUpdate.DeltaTime, 0)
	} else if !p.IsOnGround {
		// keep moving in the direction of the attack
		dx = kinematic.Displacement(p.Velocity.X, clientPlayerUpdate.DeltaTime, 0)
		vx = kinematic.FinalVelocity(p.Velocity.X, clientPlayerUpdate.DeltaTime, 0)
	}

	// Check for collisions
	if collision := p.Object.Check(dx, 0, CollisionSpaceTagLevel); collision != nil {
		dx = collision.ContactWithObject(collision.Objects[0]).X
		vx = 0
	}

	// Y-axis
	// Apply input
	vy := p.Velocity.Y
	if !p.IsAttacking && p.IsOnGround && clientPlayerUpdate.InputJump {
		vy = constants.PlayerJumpSpeed
	}

	// Apply gravity
	dy := kinematic.Displacement(vy, clientPlayerUpdate.DeltaTime, kinematic.Gravity*constants.PlayerGravityMultiplier)
	vy = kinematic.FinalVelocity(vy, clientPlayerUpdate.DeltaTime, kinematic.Gravity*constants.PlayerGravityMultiplier)

	// Check for collisions
	isOnGround := false
	if collision := p.Object.Check(0, dy, CollisionSpaceTagLevel); collision != nil {
		dy = collision.ContactWithObject(collision.Objects[0]).Y
		vy = 0
		isOnGround = true
	}

	// Update player state
	p.LastProcessedTimestamp = clientPlayerUpdate.Timestamp
	p.Position.X += dx
	p.Velocity.X = vx
	p.Position.Y += dy
	p.Velocity.Y = vy
	p.IsOnGround = isOnGround

	// Update the player animation
	if !p.IsAttacking {
		if clientPlayerUpdate.InputX > 0 {
			p.AnimationFlip = false
		} else if clientPlayerUpdate.InputX < 0 {
			p.AnimationFlip = true
		}
	}

	// Update the player collision object
	p.Object.Position.X = p.Position.X
	p.Object.Position.Y = p.Position.Y
	p.Object.Update()

	// Animation

	if p.IsAttacking {
		p.Animation = PlayerAnimationAttack
	} else {
		if isOnGround {
			if clientPlayerUpdate.InputX != 0 {
				p.Animation = PlayerAnimationRun
			} else {
				p.Animation = PlayerAnimationIdle
			}
		} else {
			if vy < 0 {
				p.Animation = PlayerAnimationJump
			} else {
				p.Animation = PlayerAnimationFall
			}
		}
	}
}
