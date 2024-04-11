package types

import (
	"github.com/cbodonnell/flywheel/pkg/game/constants"
	"github.com/cbodonnell/flywheel/pkg/kinematic"
	"github.com/solarlune/resolv"
)

type NPCState struct {
	Position      kinematic.Vector
	Velocity      kinematic.Vector
	Object        *resolv.Object
	IsOnGround    bool
	Animation     NPCAnimation
	AnimationFlip bool

	ttl         float64
	exists      bool
	respawnTime float64
}

type NPCAnimation uint8

const (
	NPCAnimationIdle NPCAnimation = iota
	// NPCAnimationRun
	// NPCAnimationJump
	// NPCAnimationFall
)

func NewNPCState(positionX float64, positionY float64) *NPCState {
	return &NPCState{
		Position: kinematic.Vector{
			X: positionX,
			Y: positionY,
		},
		Velocity: kinematic.Vector{
			X: 0,
			Y: 0,
		},
		Object: resolv.NewObject(positionX, positionY, constants.NPCWidth, constants.NPCHeight, CollisionSpaceTagNPC),
	}
}

func (n *NPCState) Copy() *NPCState {
	return &NPCState{
		Position:   n.Position,
		Velocity:   n.Velocity,
		IsOnGround: n.IsOnGround,
	}
}

func (n *NPCState) TTL() float64 {
	return n.ttl
}

func (n *NPCState) Exists() bool {
	return n.exists
}

func (n *NPCState) RespawnTime() float64 {
	return n.respawnTime
}

// Update updates the NPC state based on the current state and the time passed
func (n *NPCState) Update(deltaTime float64) {
	// X-axis
	// TODO: some base movement logic
	dx := 0.0
	vx := 0.0

	// Check for collisions
	if collision := n.Object.Check(dx, 0, CollisionSpaceTagLevel); collision != nil {
		dx = collision.ContactWithObject(collision.Objects[0]).X
		vx = 0
	}

	// Y-axis
	// TODO: some base movement logic
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

	n.Animation = NPCAnimationIdle

	// Update the npc collision object
	n.Object.Position.X = n.Position.X
	n.Object.Position.Y = n.Position.Y
	n.Object.Update()

	// Decrement the time to live
	n.ttl -= deltaTime
}

func (n *NPCState) DecrementRespawnTime(deltaTime float64) {
	n.respawnTime -= deltaTime
}

func (n *NPCState) Spawn() {
	n.ttl = constants.NPCTTL
	n.exists = true
	n.respawnTime = 0

	n.Position.X = constants.NPCStartingX
	n.Position.Y = constants.NPCStartingY
	n.Velocity.X = 0
	n.Velocity.Y = 0
	n.IsOnGround = false

	n.Object.Position.X = n.Position.X
	n.Object.Position.Y = n.Position.Y
	n.Object.Update()
}

func (n *NPCState) Despawn() {
	n.ttl = 0
	n.exists = false
	n.respawnTime = constants.NPCRespawnTime
}
