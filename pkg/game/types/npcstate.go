package types

import (
	"github.com/cbodonnell/flywheel/pkg/game/constants"
	"github.com/cbodonnell/flywheel/pkg/kinematic"
	"github.com/solarlune/resolv"
)

type NPCState struct {
	SpawnPosition kinematic.Vector
	Position      kinematic.Vector
	Velocity      kinematic.Vector
	Object        *resolv.Object
	IsOnGround    bool
	Animation     NPCAnimation
	AnimationFlip bool
	Hitpoints     int16

	exists      bool
	respawnTime float64
}

type NPCAnimation uint8

const (
	NPCAnimationIdle NPCAnimation = iota
	NPCAnimationDead
)

func NewNPCState(positionX float64, positionY float64) *NPCState {
	spawnPosition := kinematic.Vector{
		X: positionX,
		Y: positionY,
	}

	return &NPCState{
		SpawnPosition: spawnPosition,
		Object:        resolv.NewObject(positionX, positionY, constants.NPCWidth, constants.NPCHeight, CollisionSpaceTagNPC),
		Hitpoints:     constants.NPCHitpoints,
	}
}

func (n *NPCState) Copy() *NPCState {
	return &NPCState{
		Position:      n.Position,
		Velocity:      n.Velocity,
		IsOnGround:    n.IsOnGround,
		Animation:     n.Animation,
		AnimationFlip: n.AnimationFlip,
		Hitpoints:     n.Hitpoints,
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
	// TODO: some base movement logic

	// X-axis
	dx := 0.0
	vx := 0.0

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

	// Update player state
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

	if n.IsDead() {
		n.Animation = NPCAnimationDead
	} else {
		n.Animation = NPCAnimationIdle
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

	n.Object.Position.X = n.Position.X
	n.Object.Position.Y = n.Position.Y
	n.Object.Update()
}

func (n *NPCState) Despawn() {
	n.exists = false
	n.respawnTime = constants.NPCRespawnTime
}
