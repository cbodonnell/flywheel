package fonts

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

func init() {
	if err := loadFonts(); err != nil {
		panic(fmt.Sprintf("Failed to load fonts: %v", err))
	}
}

var MPlusNormalFont font.Face

func loadFonts() error {
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		return fmt.Errorf("failed to parse font: %v", err)
	}
	const dpi = 72
	MPlusNormalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingVertical,
	})
	if err != nil {
		return fmt.Errorf("failed to create font face: %v", err)
	}

	return nil
}
