package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// IsPositiveJustPressed returns a boolean value indicating whether the generic positive input is just pressed.
// This is used to handle both keyboard and touch inputs.
func IsPositiveJustPressed() bool {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return true
	}
	touchIDs := inpututil.AppendJustPressedTouchIDs(nil)
	if len(touchIDs) > 0 {
		return true
	}
	gamepadIDs := ebiten.AppendGamepadIDs(nil)
	for _, g := range gamepadIDs {
		if ebiten.IsStandardGamepadLayoutAvailable(g) {
			if inpututil.IsStandardGamepadButtonJustPressed(g, ebiten.StandardGamepadButtonRightBottom) {
				return true
			}
			if inpututil.IsStandardGamepadButtonJustPressed(g, ebiten.StandardGamepadButtonRightRight) {
				return true
			}
		} else {
			// The button 0/1 might not be A/B buttons.
			if inpututil.IsGamepadButtonJustPressed(g, ebiten.GamepadButton0) {
				return true
			}
			if inpututil.IsGamepadButtonJustPressed(g, ebiten.GamepadButton1) {
				return true
			}
		}
	}
	return false
}

// IsNegativeJustPressed returns a boolean value indicating whether the generic negative input is just pressed.
// This is used to handle both keyboard and touch inputs.
func IsNegativeJustPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEscape)
}

func IsRightPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyRight)
}

func IsLeftPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyLeft)
}

func IsUpPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyUp)
}

func IsDownPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyDown)
}

func IsJumpJustPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeySpace)
}

func IsAttackJustPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyQ)
}
