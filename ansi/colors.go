package ansi

import (
	"image/color"
	"math"
)

// AnsiPalette is the standard 16-color ANSI palette.
var AnsiPalette = []color.RGBA{
	{0x00, 0x00, 0x00, 0xff}, // 0 - Black
	{0xaa, 0x00, 0x00, 0xff}, // 1 - Red
	{0x00, 0xaa, 0x00, 0xff}, // 2 - Green
	{0xaa, 0x55, 0x00, 0xff}, // 3 - Yellow
	{0x00, 0x00, 0xaa, 0xff}, // 4 - Blue
	{0xaa, 0x00, 0xaa, 0xff}, // 5 - Magenta
	{0x00, 0xaa, 0xaa, 0xff}, // 6 - Cyan
	{0xaa, 0xaa, 0xaa, 0xff}, // 7 - White
	{0x55, 0x55, 0x55, 0xff}, // 8 - Bright Black (Gray)
	{0xff, 0x55, 0x55, 0xff}, // 9 - Bright Red
	{0x55, 0xff, 0x55, 0xff}, // 10 - Bright Green
	{0xff, 0xff, 0x55, 0xff}, // 11 - Bright Yellow
	{0x55, 0x55, 0xff, 0xff}, // 12 - Bright Blue
	{0xff, 0x55, 0xff, 0xff}, // 13 - Bright Magenta
	{0x55, 0xff, 0xff, 0xff}, // 14 - Bright Cyan
	{0xff, 0xff, 0xff, 0xff}, // 15 - Bright White
}

// colorDistance calculates the Euclidean distance between two colors.
func colorDistance(c1, c2 color.RGBA) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	rd := float64(r1) - float64(r2)
	gd := float64(g1) - float64(g2)
	bd := float64(b1) - float64(b2)
	return math.Sqrt(rd*rd + gd*gd + bd*bd)
}

// FindClosestAnsiColor finds the closest color in the 16-color ANSI palette.
// It returns the color and its index in the palette.
func FindClosestAnsiColor(c color.RGBA) (color.RGBA, int) {
	if c.A == 0 {
		return AnsiPalette[0], 0 // Transparent is black
	}

	closestIndex := 0
	minDist := math.MaxFloat64

	for i, ansiColor := range AnsiPalette {
		dist := colorDistance(c, ansiColor)
		if dist < minDist {
			minDist = dist
			closestIndex = i
		}
	}
	return AnsiPalette[closestIndex], closestIndex
}
