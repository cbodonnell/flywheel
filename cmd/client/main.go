package main

import (
	"fmt"
	"image/color"
	"log"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type GameMode int

const (
	GameModeMenu GameMode = iota
	GameModePlay
	GameModeOver
)

func (m GameMode) String() string {
	switch m {
	case GameModeMenu:
		return "Menu"
	case GameModePlay:
		return "Play"
	case GameModeOver:
		return "Over"
	}
	return "Unknown"
}

// Game implements ebiten.Game interface, which has Update, Draw and Layout methods.
type Game struct {
	// mode is the current game mode.
	mode GameMode
	// touchIDs is the last touch identifiers.
	touchIDs []ebiten.TouchID
	// gamepadIDs is the last gamepad identifiers.
	gamepadIDs []ebiten.GamepadID
}

func NewGame() ebiten.Game {
	return &Game{
		mode: GameModeMenu,
	}
}

var mplusNormalFont font.Face

func init() {
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}
	const dpi = 72
	mplusNormalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingVertical,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (g *Game) isKeyJustPressed() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return true
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return true
	}
	g.touchIDs = inpututil.AppendJustPressedTouchIDs(g.touchIDs[:0])
	if len(g.touchIDs) > 0 {
		return true
	}
	g.gamepadIDs = ebiten.AppendGamepadIDs(g.gamepadIDs[:0])
	for _, g := range g.gamepadIDs {
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

func (g *Game) Update() error {
	switch g.mode {
	case GameModeMenu:
		if g.isKeyJustPressed() {
			g.mode = GameModePlay
		}
	case GameModePlay:
		if g.isKeyJustPressed() {
			g.mode = GameModeOver
		}
	case GameModeOver:
		if g.isKeyJustPressed() {
			g.mode = GameModeMenu
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %0.2f", ebiten.ActualFPS()))
	var t string
	switch g.mode {
	case GameModeMenu:
		t = "Press to Start!"
	case GameModePlay:
		t = "Playing..."
	case GameModeOver:
		t = "Game Over"
	}
	t = strings.ToUpper(t)
	bounds, _ := font.BoundString(mplusNormalFont, t)
	op := &ebiten.DrawImageOptions{}
	// // Center the text
	op.GeoM.Translate(ScreenWidth/2-float64(bounds.Max.X>>6)/2, ScreenHeight/2-float64(bounds.Max.Y>>6)/2)
	op.ColorScale.ScaleWithColor(color.White)
	text.DrawWithOptions(screen, t, mplusNormalFont, op)
}

const (
	ScreenWidth  = 320
	ScreenHeight = 240
)

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, World!")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
