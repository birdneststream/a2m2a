package mirc

import (
	"a2m2a/canvas"
	"fmt"
	clr "image/color"
	"io"
	"math"
)

// ANSI colors to mIRC color map.
var (
	// For normal (non-bold) text
	color = []int{1, 5, 3, 7, 2, 6, 10, 15}
	// For bold text
	colorBold = []int{14, 4, 9, 8, 12, 13, 11, 0}
)

// Writer converts a canvas to a mIRC formatted string.
type Writer struct {
	canvas *canvas.Canvas
	writer io.Writer
}

// NewWriter creates a new mIRC writer.
func NewWriter(c *canvas.Canvas, w io.Writer) *Writer {
	return &Writer{
		canvas: c,
		writer: w,
	}
}

// Write generates the mIRC output from the canvas.
func (w *Writer) Write() error {
	var prevFg, prevBg clr.RGBA = canvas.DefaultFg, canvas.DefaultBg
	var prevBold bool

	// Get content bounds to treat the canvas as a fixed-size rectangle.
	// This ensures that alignment is preserved across all lines.
	_, maxRow, _, maxCol := w.canvas.GetContentBounds()

	for r, row := range w.canvas.Grid {
		if r > maxRow {
			break // Don't write trailing empty lines past the content.
		}
		// Reset state for each new line
		prevFg, prevBg = canvas.DefaultFg, canvas.DefaultBg
		prevBold = false

		for i := 0; i <= maxCol; i++ {
			cell := row[i]

			// Handle Bold state change with ^B (0x02)
			if cell.Bold != prevBold {
				fmt.Fprint(w.writer, "\x02")
				prevBold = cell.Bold
			}

			// Handle Color state change with ^C (0x03)
			if cell.Fg != prevFg || cell.Bg != prevBg {
				fgIndex, _ := findClosestMircColor(cell.Fg)
				bgIndex, _ := findClosestMircColor(cell.Bg)

				// Always write FG. Only write BG if it's not the default.
				if bgIndex != 1 { // mIRC default BG is black (index 1)
					fmt.Fprintf(w.writer, "\x03%d,%d", fgIndex, bgIndex)
				} else {
					fmt.Fprintf(w.writer, "\x03%d", fgIndex)
				}
				prevFg = cell.Fg
				prevBg = cell.Bg
			}

			if _, err := w.writer.Write([]byte(string(cell.Char))); err != nil {
				return err
			}
		}
		// mIRC doesn't need ^O to reset, a newline is enough.
		fmt.Fprint(w.writer, "\n")
	}

	return nil
}

// colorDistance calculates the Euclidean distance between two colors.
func colorDistance(c1, c2 clr.RGBA) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	rd := float64(r1) - float64(r2)
	gd := float64(g1) - float64(g2)
	bd := float64(b1) - float64(b2)
	return math.Sqrt(rd*rd + gd*gd + bd*bd)
}

// findClosestMircColor finds the closest color in the 99-color mIRC palette.
func findClosestMircColor(c clr.RGBA) (int, clr.RGBA) {
	if c.A == 0 {
		// Assuming transparent should be the default background color
		return 1, MircPalette99[1]
	}
	closestIndex := 0
	minDist := math.MaxFloat64

	for i, mircColor := range MircPalette99 {
		dist := colorDistance(c, mircColor)
		if dist < minDist {
			minDist = dist
			closestIndex = i
		}
	}
	return closestIndex, MircPalette99[closestIndex]
}
