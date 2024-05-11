package types

import (
	"math/rand"

	"github.com/cbodonnell/flywheel/pkg/game/constants"
	"github.com/cbodonnell/flywheel/pkg/kinematic"
	"github.com/solarlune/resolv"
)

type NPCState struct {
	// TODO: assess public/private fields
	ID                uint32
	SpawnPosition     kinematic.Vector
	SpawnFlip         bool
	Position          kinematic.Vector
	Velocity          kinematic.Vector
	Object            *resolv.Object
	IsOnGround        bool
	Animation         NPCAnimation
	AnimationFlip     bool
	AnimationSequence uint8
	ResetAnimation    bool
	Hitpoints         int16

	respawnTime float64

	mode         NPCMode
	FollowTarget *PlayerState

	IsInAttackRange bool
	IsAttacking     bool
	CurrentAttack   NPCAttack
	AttackTimeLeft  float64
	IsAttackHitting bool
	DidAttackHit    bool
}

type NPCAnimation uint8

const (
	NPCAnimationIdle NPCAnimation = iota
	NPCAnimationWalk
	NPCAnimationDead
	NPCAnimationAttack1
	NPCAnimationAttack2
	NPCAnimationAttack3
)

type NPCMode uint8

const (
	NPCModeIdle NPCMode = iota
	NPCModeFollow
)

type NPCAttack uint8

const (
	NPCAttack1 NPCAttack = iota
	NPCAttack2
	NPCAttack3
)

func NewNPCState(id uint32, positionX float64, positionY float64, flip bool) *NPCState {
	spawnPosition := kinematic.Vector{
		X: positionX,
		Y: positionY,
	}

	object := resolv.NewObject(positionX, positionY, constants.NPCWidth, constants.NPCHeight, CollisionSpaceTagNPC)
	object.SetShape(resolv.NewRectangle(0, 0, constants.NPCWidth, constants.NPCHeight))

	return &NPCState{
		ID:            id,
		SpawnPosition: spawnPosition,
		SpawnFlip:     flip,
		Object:        object,
	}
}

func (n *NPCState) Copy() *NPCState {
	return &NPCState{
		Position:          n.Position,
		Velocity:          n.Velocity,
		IsOnGround:        n.IsOnGround,
		IsAttacking:       n.IsAttacking,
		CurrentAttack:     n.CurrentAttack,
		AttackTimeLeft:    n.AttackTimeLeft,
		IsAttackHitting:   n.IsAttackHitting,
		DidAttackHit:      n.DidAttackHit,
		Animation:         n.Animation,
		AnimationFlip:     n.AnimationFlip,
		AnimationSequence: n.AnimationSequence,
		Hitpoints:         n.Hitpoints,
	}
}

func (n *NPCState) RespawnTime() float64 {
	return n.respawnTime
}

// Update updates the NPC state based on the current state and the time passed
// and returns whether the state has changed
func (n *NPCState) Update(deltaTime float64) (changed bool) {
	// Respawn
	if n.IsDead() {
		if n.RespawnTime() <= 0 {
			n.Spawn()
		} else {
			n.DecrementRespawnTime(deltaTime)
		}
	}

	// Attack - TODO: roll this into some kind of attack manager
	if n.IsAttacking {
		beforeIsAttacking := n.IsAttacking

		if n.AttackTimeLeft > 0 {
			n.AttackTimeLeft -= deltaTime
			if !n.DidAttackHit {
				attackHitTime := 0.0
				switch n.CurrentAttack {
				case NPCAttack1:
					attackHitTime = constants.NPCAttack1Duration - constants.NPCAttack1ChannelTime
				case NPCAttack2:
					attackHitTime = constants.NPCAttack2Duration - constants.NPCAttack2ChannelTime
				case NPCAttack3:
					attackHitTime = constants.NPCAttack3Duration - constants.NPCAttack3ChannelTime
				}

				if n.AttackTimeLeft <= attackHitTime {
					// register the hit only once
					n.IsAttackHitting = true
					n.DidAttackHit = true
				}
			} else {
				n.IsAttackHitting = false
			}
		} else {
			n.IsAttacking = false
			n.IsAttackHitting = false
			n.DidAttackHit = false
		}

		// Reset the animation sequence if the player is no longer attacking
		if beforeIsAttacking && !n.IsAttacking {
			n.ResetAnimation = true
		}
	}

	// Following
	n.UpdateFollowing()

	if !n.IsAttacking && n.IsInAttackRange {
		n.IsAttacking = true
		// randomly choose an attack
		attack := rand.Intn(3)
		switch attack {
		case 0:
			n.CurrentAttack = NPCAttack1
			n.AttackTimeLeft = constants.NPCAttack1Duration
		case 1:
			n.CurrentAttack = NPCAttack2
			n.AttackTimeLeft = constants.NPCAttack2Duration
		case 2:
			n.CurrentAttack = NPCAttack3
			n.AttackTimeLeft = constants.NPCAttack3Duration
		}
	}

	// Movement

	// X-axis
	dx := 0.0
	vx := 0.0
	if !n.IsAttacking {
		if n.IsFollowing() {
			if n.FollowTarget.Position.X < n.Position.X {
				vx = -constants.NPCSpeed
			} else if n.FollowTarget.Position.X > n.Position.X {
				vx = constants.NPCSpeed
			}
			dx = kinematic.Displacement(vx, deltaTime, 0)
			vx = kinematic.FinalVelocity(vx, deltaTime, 0)

			if n.FollowTarget.Position.X < n.Position.X && n.Position.X+dx < n.FollowTarget.Position.X ||
				n.FollowTarget.Position.X > n.Position.X && n.Position.X+dx > n.FollowTarget.Position.X {
				// handle edge case where npc is directly on top of player and oscillates
				dx = 0
			}
		}
		// TODO: else wander
	}

	// Check for collisions
	if collision := n.Object.Check(dx, 0, CollisionSpaceTagLevel); collision != nil {
		dx = collision.ContactWithObject(collision.Objects[0]).X
		vx = 0
	}

	// Update npc state in the X-axis
	n.Position.X += dx
	n.Velocity.X = vx
	n.Object.Position.X = n.Position.X
	n.Object.Update()

	// Y-axis
	vy := n.Velocity.Y

	// Apply gravity
	dy := kinematic.Displacement(vy, deltaTime, kinematic.Gravity*constants.NPCGravityMultiplier)
	vy = kinematic.FinalVelocity(vy, deltaTime, kinematic.Gravity*constants.NPCGravityMultiplier)

	// Check for collisions
	isOnGround := false
	if collision := n.Object.Check(0, dy, CollisionSpaceTagLevel); collision != nil {
		dy = collision.ContactWithObject(collision.Objects[0]).Y
		vy = 0
		isOnGround = true
	}

	// Update npc state in the Y-axis
	n.Position.Y += dy
	n.Velocity.Y = vy
	n.Object.Position.Y = n.Position.Y
	n.Object.Update()
	n.IsOnGround = isOnGround

	// Update the npc animation
	if n.Velocity.X > 0 {
		n.AnimationFlip = false
	} else if n.Velocity.X < 0 {
		n.AnimationFlip = true
	}

	// Animation
	beforeAnimation := n.Animation

	if n.IsAttacking {
		switch n.CurrentAttack {
		case NPCAttack1:
			n.Animation = NPCAnimationAttack1
		case NPCAttack2:
			n.Animation = NPCAnimationAttack2
		case NPCAttack3:
			n.Animation = NPCAnimationAttack3
		}
	} else {
		if n.IsDead() {
			n.Animation = NPCAnimationDead
		} else {
			if n.mode == NPCModeFollow {
				n.Animation = NPCAnimationWalk
			} else {
				n.Animation = NPCAnimationIdle
			}
		}
	}

	// Update the animation sequence
	if beforeAnimation != n.Animation || n.ResetAnimation {
		n.AnimationSequence++
		n.ResetAnimation = false
	}

	// TODO: return false if the update did not change the state
	return true
}

func (n *NPCState) TakeDamage(damage int16) {
	n.Hitpoints -= damage
}

func (n *NPCState) IsDead() bool {
	return n.Hitpoints <= 0
}

func (n *NPCState) DecrementRespawnTime(deltaTime float64) {
	n.respawnTime -= deltaTime
}

func (n *NPCState) Spawn() {
	n.respawnTime = 0
	n.mode = NPCModeIdle
	n.FollowTarget = nil

	n.Position = kinematic.NewVector(n.SpawnPosition.X, n.SpawnPosition.Y)
	n.Velocity = kinematic.ZeroVector()
	n.IsOnGround = false

	n.Hitpoints = constants.NPCHitpoints

	n.AnimationFlip = n.SpawnFlip
	n.Animation = NPCAnimationIdle
	n.AnimationSequence = 0
	n.ResetAnimation = false

	n.IsAttacking = false
	n.CurrentAttack = NPCAttack1
	n.AttackTimeLeft = 0
	n.IsAttackHitting = false
	n.DidAttackHit = false

	n.Object.Position.X = n.Position.X
	n.Object.Position.Y = n.Position.Y
	n.Object.Update()
}

func (n *NPCState) Despawn() {
	n.respawnTime = constants.NPCRespawnTime
}

func (n *NPCState) StartFollowing(target *PlayerState) {
	n.mode = NPCModeFollow
	n.FollowTarget = target
}

func (n *NPCState) StopFollowing() {
	n.mode = NPCModeIdle
	n.FollowTarget = nil
	n.IsInAttackRange = false
}

func (n *NPCState) IsFollowing() bool {
	return n.mode == NPCModeFollow
}

func (n *NPCState) UpdateFollowing() {
	if n.FollowTarget == nil {
		n.StopFollowing()
		return
	}

	if n.IsDead() {
		n.StopFollowing()
		return
	}

	if n.FollowTarget.IsDead() {
		n.StopFollowing()
		return
	}

	// check if the npc is too far from the player
	if n.Position.DistanceFrom(n.FollowTarget.Position) > 2*constants.NPCLineOfSight {
		n.StopFollowing()
		return
	}

	// check if the npc has a direct line of sight to the player
	lineOfSight := resolv.NewLine(n.Position.X+constants.NPCWidth/2, n.Position.Y+constants.NPCHeight/2, n.FollowTarget.Position.X+constants.PlayerWidth/2, n.FollowTarget.Position.Y+constants.PlayerHeight/2)
	for _, obj := range n.Object.Space.Objects() {
		if !obj.HasTags(CollisionSpaceTagLevel) {
			continue
		}
		if contact := lineOfSight.Intersection(0, 0, obj.Shape); contact != nil {
			n.StopFollowing()
			return
		}
	}

	// check if the target is in front of the npc and within attack range
	flip := 1.0
	if n.AnimationFlip {
		flip = -1.0
	}
	xDistance := n.FollowTarget.Position.X - n.Position.X
	if flip*xDistance > 0 && flip*xDistance < constants.NPCAttackRange {
		n.IsInAttackRange = true
	} else {
		n.IsInAttackRange = false
	}
}
