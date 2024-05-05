package types

import (
	"github.com/cbodonnell/flywheel/pkg/game/constants"
	"github.com/cbodonnell/flywheel/pkg/kinematic"
	"github.com/solarlune/resolv"
)

type NPCState struct {
	SpawnPosition kinematic.Vector
	SpawnFlip     bool
	Position      kinematic.Vector
	Velocity      kinematic.Vector
	Object        *resolv.Object
	IsOnGround    bool
	Animation     NPCAnimation
	AnimationFlip bool
	Hitpoints     int16

	exists      bool
	respawnTime float64

	mode         NPCMode
	followTarget *PlayerState

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
)

type NPCMode uint8

const (
	NPCModeIdle NPCMode = iota
	NPCModeFollow
)

type NPCAttack uint8

const (
	NPCAttack1 NPCAttack = iota
	// NPCAttack2
	// NPCAttack3
)

func NewNPCState(positionX float64, positionY float64, flip bool) *NPCState {
	spawnPosition := kinematic.Vector{
		X: positionX,
		Y: positionY,
	}

	object := resolv.NewObject(positionX, positionY, constants.NPCWidth, constants.NPCHeight, CollisionSpaceTagNPC)
	object.SetShape(resolv.NewRectangle(0, 0, constants.NPCWidth, constants.NPCHeight))

	return &NPCState{
		SpawnPosition: spawnPosition,
		SpawnFlip:     flip,
		AnimationFlip: flip,
		Object:        object,
		Hitpoints:     constants.NPCHitpoints,
	}
}

func (n *NPCState) Copy() *NPCState {
	return &NPCState{
		Position:        n.Position,
		Velocity:        n.Velocity,
		IsOnGround:      n.IsOnGround,
		IsAttacking:     n.IsAttacking,
		CurrentAttack:   n.CurrentAttack,
		AttackTimeLeft:  n.AttackTimeLeft,
		IsAttackHitting: n.IsAttackHitting,
		DidAttackHit:    n.DidAttackHit,
		Animation:       n.Animation,
		AnimationFlip:   n.AnimationFlip,
		Hitpoints:       n.Hitpoints,
	}
}

func (n *NPCState) Exists() bool {
	return n.exists
}

func (n *NPCState) RespawnTime() float64 {
	return n.respawnTime
}

// Update updates the NPC state based on the current state and the time passed
// and returns whether the state has changed
func (n *NPCState) Update(deltaTime float64) (changed bool) {
	// Attack

	if n.AttackTimeLeft > 0 {
		n.AttackTimeLeft -= deltaTime
		if !n.DidAttackHit {
			attackHitTime := 0.0
			switch n.CurrentAttack {
			case NPCAttack1:
				attackHitTime = constants.NPCAttack1Duration - constants.NPCAttack1ChannelTime
				// case NPCAttack2:
				// 	attackHitTime = constants.NPCAttack2Duration - constants.NPCAttack2ChannelTime
				// case NPCAttack3:
				// 	attackHitTime = constants.NPCAttack3Duration - constants.NPCAttack3ChannelTime
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

	if !n.IsAttacking && n.IsInAttackRange {
		n.IsAttacking = true
		n.CurrentAttack = NPCAttack1 // TODO: randomize
		n.AttackTimeLeft = constants.NPCAttack1Duration
	}

	// Movement

	// X-axis
	dx := 0.0
	vx := 0.0
	if !n.IsAttacking {
		if n.mode == NPCModeFollow {
			if n.followTarget.Position.X < n.Position.X {
				vx = -constants.NPCSpeed
			} else if n.followTarget.Position.X > n.Position.X {
				vx = constants.NPCSpeed
			}
			dx = kinematic.Displacement(vx, deltaTime, 0)
			vx = kinematic.FinalVelocity(vx, deltaTime, 0)
		}
		// TODO: else wander
	}

	// Check for collisions
	if collision := n.Object.Check(dx, 0, CollisionSpaceTagLevel); collision != nil {
		dx = collision.ContactWithObject(collision.Objects[0]).X
		vx = 0
	}

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

	// Update npc state
	n.Position.X += dx
	n.Velocity.X = vx
	n.Position.Y += dy
	n.Velocity.Y = vy
	n.IsOnGround = isOnGround

	// Update the npc animation
	if n.Velocity.X > 0 {
		n.AnimationFlip = false
	} else if n.Velocity.X < 0 {
		n.AnimationFlip = true
	}

	// Update the npc collision object
	n.Object.Position.X = n.Position.X
	n.Object.Position.Y = n.Position.Y
	n.Object.Update()

	// Animation

	if n.IsAttacking {
		switch n.CurrentAttack {
		case NPCAttack1:
			n.Animation = NPCAnimationAttack1
			// case NPCAttack2:
			// 	n.Animation = NPCAnimationAttack2
			// case NPCAttack3:
			// 	n.Animation = NPCAnimationAttack3
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
	n.exists = true
	n.respawnTime = 0

	n.Position.X = n.SpawnPosition.X
	n.Position.Y = n.SpawnPosition.Y
	n.Velocity.X = 0
	n.Velocity.Y = 0
	n.IsOnGround = false

	n.Hitpoints = constants.NPCHitpoints

	n.AnimationFlip = n.SpawnFlip
	n.Animation = NPCAnimationIdle

	n.Object.Position.X = n.Position.X
	n.Object.Position.Y = n.Position.Y
	n.Object.Update()
}

func (n *NPCState) Despawn() {
	n.StopFollowing()
	n.exists = false
	n.respawnTime = constants.NPCRespawnTime
}

func (n *NPCState) StartFollowing(target *PlayerState) {
	n.mode = NPCModeFollow
	n.followTarget = target
}

func (n *NPCState) StopFollowing() {
	n.mode = NPCModeIdle
	n.followTarget = nil
}

func (n *NPCState) IsFollowing() bool {
	return n.mode == NPCModeFollow
}

func (n *NPCState) UpdateFollowing() {
	// TODO: figure out why this is not toggling off properly
	n.IsInAttackRange = false

	// check if the npc is too far from the player
	distanceFromTarget := n.Position.DistanceFrom(n.followTarget.Position)
	if distanceFromTarget > 2*constants.NPCLineOfSight {
		n.StopFollowing()
		return
	}

	// check if the npc has a direct line of sight to the player
	lineOfSight := resolv.NewLine(n.Position.X+constants.NPCWidth/2, n.Position.Y+constants.NPCHeight/2, n.followTarget.Position.X+constants.PlayerWidth/2, n.followTarget.Position.Y+constants.PlayerHeight/2)
	for _, obj := range n.Object.Space.Objects() {
		if !obj.HasTags(CollisionSpaceTagLevel) {
			continue
		}
		if contact := lineOfSight.Intersection(0, 0, obj.Shape); contact != nil {
			n.StopFollowing()
			return
		}
	}

	// check if the npc is close enough to the player to attack
	if distanceFromTarget < constants.NPCAttackRange {
		n.IsInAttackRange = true
	}
}
