package mirc

import (
	"a2m2a/canvas"
	"fmt"
	"io"
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
	var prevCell canvas.Cell

	for _, row := range w.canvas.Grid {
		// Find the last meaningful character in the row to trim trailing space.
		lastCharIndex := -1
		for i := len(row) - 1; i >= 0; i-- {
			cell := row[i]
			if cell.Char != ' ' || cell.Bg != canvas.DefaultBg {
				lastCharIndex = i
				break
			}
		}

		// If the line is empty, we don't write anything, not even a newline,
		// unless we want to preserve blank lines. For now, we'll skip them
		// to create the most compact output. If we print just "\n", some
		// clients interpret that as a color reset.
		if lastCharIndex == -1 {
			continue
		}

		// At the start of a new line, the 'previous' state is the default.
		prevCell = canvas.Cell{
			Fg:   canvas.DefaultFg,
			Bg:   canvas.DefaultBg,
			Bold: canvas.DefaultBold,
			Ice:  canvas.DefaultIce,
		}

		for i := 0; i <= lastCharIndex; i++ {
			cell := row[i]
			if cell.Fg != prevCell.Fg || cell.Bg != prevCell.Bg || cell.Bold != prevCell.Bold || cell.Ice != prevCell.Ice {
				fmt.Fprintf(w.writer, "\x03")
				fmt.Fprintf(w.writer, "%d", getFgColor(&cell))

				// The mIRC spec says to include the background color to change it.
				// If the previous char had a background and this one doesn't, we must explicitly set the default.
				// This logic is more robust than the original C code's for client compatibility.
				// We write the BG color if it's not default, or if it has changed from the previous cell's BG.
				if cell.Bg != canvas.DefaultBg || cell.Bg != prevCell.Bg || cell.Ice != prevCell.Ice {
					fmt.Fprintf(w.writer, ",%d", getBgColor(&cell))
				}
			}

			if _, err := w.writer.Write([]byte(string(cell.Char))); err != nil {
				return err
			}
			prevCell = cell
		}

		if _, err := w.writer.Write([]byte("\n")); err != nil {
			return err
		}
	}

	return nil
}

func getFgColor(cell *canvas.Cell) int {
	if cell.Bold {
		return colorBold[cell.Fg]
	}
	return color[cell.Fg]
}

func getBgColor(cell *canvas.Cell) int {
	if cell.Ice {
		// The C code uses color_bold for iCE backgrounds, let's replicate that.
		return colorBold[cell.Bg]
	}
	return color[cell.Bg]
}
