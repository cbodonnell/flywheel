package collisions

import "github.com/solarlune/resolv"

// TODO: make this dynamic
func NewCollisionSpace() *resolv.Space {
	space := resolv.NewSpace(640, 480, 16, 16)
	space.Add(
		resolv.NewObject(0, 0, 640, 16),
		resolv.NewObject(0, 480-16, 640, 16),
		resolv.NewObject(0, 16, 16, 480-32),
		resolv.NewObject(640-16, 16, 16, 480-32),
	)
	return space
}
