package out

import (
	"fmt"
	"math/rand"
)

func RandomVsibleHexColor() string {
	// Randomize R, G, B values within a mid-range (64 and 191) for better
	// contrast on light and dark backgrounds
	r := rand.Intn(128) + 64
	g := rand.Intn(128) + 64
	b := rand.Intn(128) + 64

	// Format the RGB values into hex color code
	hexColor := fmt.Sprintf("#%02X%02X%02X", r, g, b)
	return hexColor
}
