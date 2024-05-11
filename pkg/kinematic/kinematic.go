package kinematic

// This package includes functions for the big four kinematic equations.

import (
	"math"
)

const (
	Gravity float64 = -9.8
)

type Vector struct {
	X float64
	Y float64
}

func NewVector(x float64, y float64) Vector {
	return Vector{X: x, Y: y}
}

func ZeroVector() Vector {
	return Vector{X: 0, Y: 0}
}

func (v Vector) Equals(other Vector) bool {
	return v.X == other.X && v.Y == other.Y
}

func (v Vector) DistanceFrom(other Vector) float64 {
	return math.Sqrt(math.Pow(v.X-other.X, 2) + math.Pow(v.Y-other.Y, 2))
}

// Displacement returns the displacement of an object given its initial velocity, time, and acceleration.
func Displacement(initialVelocity float64, time float64, acceleration float64) float64 {
	return initialVelocity*time + 0.5*acceleration*math.Pow(time, 2)
}

// FinalVelocity returns the final velocity of an object given its initial velocity, time, and acceleration.
func FinalVelocity(initialVelocity float64, time float64, acceleration float64) float64 {
	return initialVelocity + acceleration*time
}

// MoveTowards returns the displacement of an object moving towards a target given its speed, time, and acceleration.
// The displacement is clamped to the target position.
func MoveTowards(speed float64, time float64, acceleration float64, target float64, position float64) float64 {
	// if the target is behind the object, the velocity should be negative
	initialVelocity := speed
	if target < position {
		initialVelocity = -speed
	}
	displacement := Displacement(initialVelocity, time, acceleration)
	if target < position && position+displacement < target {
		displacement = target - position
	} else if target > position && position+displacement > target {
		displacement = target - position
	}
	return displacement
}
