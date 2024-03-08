package kinematic

// This package includes functions for the big four kinematic equations.

import (
	"math"
)

const (
	Gravity float64 = -9.8
)

// Displacement returns the displacement of an object given its initial velocity, time, and acceleration.
func Displacement(initialVelocity float64, time float64, acceleration float64) float64 {
	return initialVelocity*time + 0.5*acceleration*math.Pow(time, 2)
}

// FinalVelocity returns the final velocity of an object given its initial velocity, time, and acceleration.
func FinalVelocity(initialVelocity float64, time float64, acceleration float64) float64 {
	return initialVelocity + acceleration*time
}
