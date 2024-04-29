package types

import (
	"crypto/sha256"
	"fmt"

	"github.com/cbodonnell/flywheel/pkg/game/constants"
	"github.com/cbodonnell/flywheel/pkg/kinematic"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/solarlune/resolv"
)

type PlayerState struct {
	LastProcessedTimestamp int64
	UserID                 string
	Name                   string
	Position               kinematic.Vector
	Velocity               kinematic.Vector
	Object                 *resolv.Object
	IsOnGround             bool
	IsAttacking            bool
	CurrentAttack          PlayerAttack
	AttackTimeLeft         float64
	IsAttackHitting        bool
	DidAttackHit           bool
	Animation              PlayerAnimation
	AnimationFlip          bool
}

type PlayerAttack uint8

const (
	PlayerAttack1 PlayerAttack = iota
	PlayerAttack2
	PlayerAttack3
)

type PlayerAnimation uint8

const (
	PlayerAnimationIdle PlayerAnimation = iota
	PlayerAnimationRun
	PlayerAnimationJump
	PlayerAnimationFall
	PlayerAnimationAttack1
	PlayerAnimationAttack2
	PlayerAnimationAttack3
)

func NewPlayerState(playerID string, positionX, positionY float64) *PlayerState {
	name := fmt.Sprintf("%x", sha256.Sum256([]byte(playerID)))[:8]

	return &PlayerState{
		UserID: playerID,
		Name:   name,
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
		UserID:                 p.UserID,
		Name:                   p.Name,
		Position:               p.Position,
		Velocity:               p.Velocity,
		IsOnGround:             p.IsOnGround,
		IsAttacking:            p.IsAttacking,
		Animation:              p.Animation,
		AnimationFlip:          p.AnimationFlip,
	}
}

// ApplyInput updates the player's state based on the client's input
// and returns whether the state has changed
func (p *PlayerState) ApplyInput(clientPlayerUpdate *messages.ClientPlayerUpdate) (changed bool) {
	// Attack

	if p.AttackTimeLeft > 0 {
		p.AttackTimeLeft -= clientPlayerUpdate.DeltaTime
		if !p.DidAttackHit {
			attackHitTime := 0.0
			switch p.CurrentAttack {
			case PlayerAttack1:
				attackHitTime = constants.PlayerAttack1Duration - constants.PlayerAttack1ChannelTime
			case PlayerAttack2:
				attackHitTime = constants.PlayerAttack2Duration - constants.PlayerAttack2ChannelTime
			case PlayerAttack3:
				attackHitTime = constants.PlayerAttack3Duration - constants.PlayerAttack3ChannelTime
			}

			if p.AttackTimeLeft <= attackHitTime {
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

	if !p.IsAttacking {
		switch {
		case clientPlayerUpdate.InputAttack1:
			p.IsAttacking = true
			p.CurrentAttack = PlayerAttack1
			p.AttackTimeLeft = constants.PlayerAttack1Duration
		case clientPlayerUpdate.InputAttack2:
			p.IsAttacking = true
			p.CurrentAttack = PlayerAttack2
			p.AttackTimeLeft = constants.PlayerAttack2Duration
		case clientPlayerUpdate.InputAttack3:
			p.IsAttacking = true
			p.CurrentAttack = PlayerAttack3
			p.AttackTimeLeft = constants.PlayerAttack3Duration
		}
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
		dx = collision.ContactWithCell(collision.Cells[0]).X
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
		dy = collision.ContactWithCell(collision.Cells[0]).Y
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
		switch p.CurrentAttack {
		case PlayerAttack1:
			p.Animation = PlayerAnimationAttack1
		case PlayerAttack2:
			p.Animation = PlayerAnimationAttack2
		case PlayerAttack3:
			p.Animation = PlayerAnimationAttack3
		}
	} else {
		if isOnGround {
			if clientPlayerUpdate.InputX != 0 {
				p.Animation = PlayerAnimationRun
			} else {
				p.Animation = PlayerAnimationIdle
			}
		} else {
			if vy < 0 {
				p.Animation = PlayerAnimationFall
			} else {
				p.Animation = PlayerAnimationJump
			}
		}
	}

	// TODO: return false if the update did not change the state
	return true
}
