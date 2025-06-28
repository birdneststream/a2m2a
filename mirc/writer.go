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

	// Get content bounds to treat the canvas as a fixed-size rectangle.
	// This ensures that alignment is preserved across all lines.
	_, maxRow, _, maxCol := w.canvas.GetContentBounds()

	for r, row := range w.canvas.Grid {
		if r > maxRow {
			break // Don't write trailing empty lines past the content.
		}

		// At the start of a new line, the 'previous' state is the default.
		prevCell = canvas.Cell{
			Fg:     canvas.DefaultFg,
			Bg:     canvas.DefaultBg,
			Bold:   canvas.DefaultBold,
			Bright: false,
			Ice:    canvas.DefaultIce,
		}

		for i := 0; i <= maxCol; i++ {
			cell := row[i]

			// Handle Bold state change with ^B (0x02)
			if cell.Bold != prevCell.Bold {
				fmt.Fprint(w.writer, "\x02")
			}

			// Handle Color state change with ^C (0x03)
			if cell.Bright != prevCell.Bright || cell.Fg != prevCell.Fg || cell.Bg != prevCell.Bg || cell.Ice != prevCell.Ice {
				fmt.Fprintf(w.writer, "\x03")
				fmt.Fprintf(w.writer, "%d", getFgColor(&cell))

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
		fmt.Fprint(w.writer, "\n")
	}

	return nil
}

func getFgColor(cell *canvas.Cell) int {
	if cell.Bright {
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
