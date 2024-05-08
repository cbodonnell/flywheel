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
