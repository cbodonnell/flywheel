package types

import (
	"github.com/cbodonnell/flywheel/pkg/game/constants"
	"github.com/cbodonnell/flywheel/pkg/kinematic"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/solarlune/resolv"
)

type PlayerState struct {
	LastProcessedTimestamp int64
	CharacterID            int32
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
	AnimationSequence      uint8
	ResetAnimation         bool
	Hitpoints              int16
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
	PlayerAnimationDead
)

func NewPlayerState(characterID int32, name string, position kinematic.Vector, hitpoints int16) *PlayerState {
	object := resolv.NewObject(position.X, position.Y, constants.PlayerWidth, constants.PlayerHeight, CollisionSpaceTagPlayer)
	object.SetShape(resolv.NewRectangle(0, 0, constants.PlayerWidth, constants.PlayerHeight))

	return &PlayerState{
		CharacterID: characterID,
		Name:        name,
		Position:    position,
		Velocity: kinematic.Vector{
			X: 0,
			Y: 0,
		},
		Hitpoints: hitpoints,
		Object:    object,
	}
}

// Equals returns true if the player state is equal to another player state.
// It only compares the fields that are serialized to the client.
func (p *PlayerState) Equals(other *PlayerState) bool {
	return p.Position.Equals(other.Position) &&
		p.Velocity.Equals(other.Velocity) &&
		p.IsOnGround == other.IsOnGround &&
		p.IsAttacking == other.IsAttacking &&
		p.Animation == other.Animation &&
		p.AnimationFlip == other.AnimationFlip &&
		p.AnimationSequence == other.AnimationSequence &&
		p.Hitpoints == other.Hitpoints
}

// Copy returns a copy of the player state with an empty object reference
func (p *PlayerState) Copy() *PlayerState {
	return &PlayerState{
		LastProcessedTimestamp: p.LastProcessedTimestamp,
		CharacterID:            p.CharacterID,
		Name:                   p.Name,
		Position:               p.Position,
		Velocity:               p.Velocity,
		IsOnGround:             p.IsOnGround,
		IsAttacking:            p.IsAttacking,
		CurrentAttack:          p.CurrentAttack,
		AttackTimeLeft:         p.AttackTimeLeft,
		IsAttackHitting:        p.IsAttackHitting,
		DidAttackHit:           p.DidAttackHit,
		Animation:              p.Animation,
		AnimationFlip:          p.AnimationFlip,
		AnimationSequence:      p.AnimationSequence,
		Hitpoints:              p.Hitpoints,
	}
}

// ApplyInput updates the player's state based on the client's input
// and returns whether the state has changed
func (p *PlayerState) ApplyInput(clientPlayerUpdate *messages.ClientPlayerUpdate) (changed bool) {
	previousState := p.Copy()

	p.LastProcessedTimestamp = clientPlayerUpdate.Timestamp

	// Respawn
	if p.IsDead() && clientPlayerUpdate.InputRespawn {
		p.Respawn(kinematic.NewVector(constants.PlayerStartingX, constants.PlayerStartingY))
	}

	// Attack - TODO: roll this into some kind of attack manager
	if p.IsAttacking {
		beforeIsAttacking := p.IsAttacking

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

		// Reset the animation sequence if the player is no longer attacking
		if beforeIsAttacking && !p.IsAttacking {
			p.ResetAnimation = true
		}
	}

	if !p.IsAttacking && !p.IsDead() {
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
	if !p.IsAttacking && !p.IsDead() {
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

	// Update player state in the X-axis
	p.Position.X += dx
	p.Velocity.X = vx
	p.Object.Position.X = p.Position.X
	p.Object.Update()

	// Y-axis
	// Apply input
	vy := p.Velocity.Y
	if !p.IsAttacking && !p.IsDead() && p.IsOnGround && clientPlayerUpdate.InputJump {
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

	// Update player state in the Y-axis
	p.Position.Y += dy
	p.Velocity.Y = vy
	p.Object.Position.Y = p.Position.Y
	p.Object.Update()
	p.IsOnGround = isOnGround

	// Update the player animation
	if !p.IsAttacking && !p.IsDead() {
		if clientPlayerUpdate.InputX > 0 {
			p.AnimationFlip = false
		} else if clientPlayerUpdate.InputX < 0 {
			p.AnimationFlip = true
		}
	}

	// Animation
	beforeAnimation := p.Animation

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
		if p.IsDead() {
			p.Animation = PlayerAnimationDead
		} else {
			if p.IsOnGround {
				if p.Velocity.X != 0 {
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
	}

	// Update the player animation sequence
	if beforeAnimation != p.Animation || p.ResetAnimation {
		p.AnimationSequence++
		p.ResetAnimation = false
	}

	return !p.Equals(previousState)
}

// TakeDamage reduces the player's hitpoints by the given amount
func (p *PlayerState) TakeDamage(damage int16) {
	p.Hitpoints -= damage
}

// IsDead returns true if the player's hitpoints are less than or equal to zero
func (p *PlayerState) IsDead() bool {
	return p.Hitpoints <= 0
}

// Respawns the player at the given position
func (p *PlayerState) Respawn(position kinematic.Vector) {
	p.Position = position
	p.Velocity = kinematic.ZeroVector()
	p.Hitpoints = constants.PlayerHitpoints

	p.IsAttacking = false
	p.CurrentAttack = PlayerAttack1
	p.AttackTimeLeft = 0
	p.IsAttackHitting = false
	p.DidAttackHit = false

	p.Animation = PlayerAnimationIdle
	p.AnimationFlip = false
	p.AnimationSequence = 0
	p.ResetAnimation = false

	p.Object.Position.X = p.Position.X
	p.Object.Position.Y = p.Position.Y
	p.Object.Update()
}
