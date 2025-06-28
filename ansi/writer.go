package ansi

import (
	"a2m2a/canvas"
	"fmt"
	"io"
	"strings"
)

// Writer converts a canvas to an ANSI formatted string.
type Writer struct {
	canvas *canvas.Canvas
	writer io.Writer
}

// NewWriter creates a new ANSI writer.
func NewWriter(c *canvas.Canvas, w io.Writer) *Writer {
	return &Writer{
		canvas: c,
		writer: w,
	}
}

// Write generates the ANSI output from the canvas.
func (w *Writer) Write() error {
	var prevCell canvas.Cell
	// Initialize prevCell to defaults so the first cell is always rendered.
	prevCell = canvas.Cell{
		Fg:     canvas.DefaultFg,
		Bg:     canvas.DefaultBg,
		Bold:   canvas.DefaultBold,
		Bright: false,
		Ice:    canvas.DefaultIce,
	}

	// Get content bounds to avoid writing trailing spaces for the entire canvas width
	_, maxRow, _, _ := w.canvas.GetContentBounds()

	for r := 0; r <= maxRow; r++ {
		row := w.canvas.Grid[r]
		// Find the last non-space character to trim trailing spaces.
		lastCharIndex := -1
		for i := len(row) - 1; i >= 0; i-- {
			if row[i].Char != ' ' || row[i].Bg != canvas.DefaultBg {
				lastCharIndex = i
				break
			}
		}

		if lastCharIndex == -1 {
			fmt.Fprint(w.writer, "\n")
			continue
		}

		for i := 0; i <= lastCharIndex; i++ {
			cell := row[i]
			if cell != prevCell {
				var params []string
				// Reset is code 0, which also conveniently resets bold/ice.
				params = append(params, "0")

				// Set bold if needed. Brightness is handled by the color code.
				if cell.Bold {
					params = append(params, "1")
				}

				if cell.Bright {
					params = append(params, fmt.Sprintf("%d", cell.Fg+90))
				} else {
					params = append(params, fmt.Sprintf("%d", cell.Fg+30))
				}

				// The original C code used blink (5) for iCE colors.
				if cell.Ice {
					params = append(params, "5")
					params = append(params, fmt.Sprintf("%d", cell.Bg+100))
				} else {
					params = append(params, fmt.Sprintf("%d", cell.Bg+40))
				}

				fmt.Fprintf(w.writer, "\x1b[%sm", strings.Join(params, ";"))
			}
			if _, err := w.writer.Write([]byte(string(cell.Char))); err != nil {
				return err
			}
			prevCell = cell
		}
		// Reset attributes at the end of the line.
		fmt.Fprint(w.writer, "\x1b[0m\n")
	}

	return nil
}
