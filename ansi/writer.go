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
	// To begin, reset all attributes
	if _, err := fmt.Fprint(w.writer, "\x1b[0m"); err != nil {
		return err
	}
	// Move cursor to top-left
	if _, err := fmt.Fprint(w.writer, "\x1b[H"); err != nil {
		return err
	}

	for _, row := range w.canvas.Grid {
		// At the start of each line, the previous cell state should be reset to default,
		// because the prior line ended with a reset code.
		var prevCell canvas.Cell
		prevCell = canvas.Cell{
			Fg:   canvas.DefaultFg,
			Bg:   canvas.DefaultBg,
			Bold: canvas.DefaultBold,
			Ice:  canvas.DefaultIce,
		}

		// Find the last meaningful character in the row to trim trailing space.
		lastCharIndex := -1
		for i := len(row) - 1; i >= 0; i-- {
			cell := row[i]
			// A cell is meaningful if it's not a space with a default background.
			if cell.Char != ' ' || cell.Bg != canvas.DefaultBg {
				lastCharIndex = i
				break
			}
		}

		// If the line is empty (all default spaces), just print a newline.
		if lastCharIndex == -1 {
			fmt.Fprint(w.writer, "\n")
			continue
		}

		for i := 0; i <= lastCharIndex; i++ {
			cell := row[i]
			if cell.Fg != prevCell.Fg || cell.Bg != prevCell.Bg || cell.Bold != prevCell.Bold || cell.Ice != prevCell.Ice {
				params := []string{}

				// A single SGR sequence can set multiple attributes.
				// We build a list of parameters and join them.
				// Example: \x1b[1;31;44m for bold, red fg, blue bg.

				// It's most reliable to reset and set all attributes for the cell.
				params = append(params, "0")

				fgCode := cell.Fg + 30
				if cell.Bold {
					fgCode = cell.Fg + 90
				}
				params = append(params, fmt.Sprintf("%d", fgCode))

				bgCode := cell.Bg + 40
				if cell.Ice {
					bgCode = cell.Bg + 100
				}
				// We only need to specify the background if it's not the default.
				if bgCode != 40 {
					params = append(params, fmt.Sprintf("%d", bgCode))
				}

				fmt.Fprintf(w.writer, "\x1b[%sm", strings.Join(params, ";"))
			}

			if _, err := w.writer.Write([]byte(string(cell.Char))); err != nil {
				return err
			}
			prevCell = cell
		}
		// At the end of a row, reset attributes and print a newline for safety.
		fmt.Fprint(w.writer, "\x1b[0m\n")
	}

	// Reset attributes at the very end.
	fmt.Fprint(w.writer, "\x1b[0m")
	return nil
}
