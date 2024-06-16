package fonts

import (
	"fmt"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

func init() {
	if err := loadFonts(); err != nil {
		panic(fmt.Sprintf("Failed to load fonts: %v", err))
	}
}

var TTFLargeFont font.Face
var TTFNormalFont font.Face
var TTFSmallFont font.Face
var TTFTinyFont font.Face

func loadFonts() error {
	const dpi = 72
	ttfFont, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return fmt.Errorf("failed to parse font: %v", err)
	}

	TTFLargeFont = truetype.NewFace(ttfFont, &truetype.Options{
		Size:    32,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})

	TTFNormalFont = truetype.NewFace(ttfFont, &truetype.Options{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})

	TTFSmallFont = truetype.NewFace(ttfFont, &truetype.Options{
		Size:    12,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})

	TTFTinyFont = truetype.NewFace(ttfFont, &truetype.Options{
		Size:    8,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})

	return nil
}
