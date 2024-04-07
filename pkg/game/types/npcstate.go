package types

import (
	"github.com/cbodonnell/flywheel/pkg/game/constants"
	"github.com/cbodonnell/flywheel/pkg/kinematic"
	"github.com/solarlune/resolv"
)

type NPCState struct {
	Position   kinematic.Vector
	Velocity   kinematic.Vector
	Object     *resolv.Object
	IsOnGround bool
	// Animation     PlayerAnimation
	// AnimationFlip bool
}

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
	vy := 0.0

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

	// TODO: Update the npc animation

	n.Object.Position.X = n.Position.X
	n.Object.Position.Y = n.Position.Y
	n.Object.Update()
}
