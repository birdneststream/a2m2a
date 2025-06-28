package ansi

import (
	"a2m2a/canvas"
	"fmt"
	"image/color"
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
	var prevFg, prevBg color.RGBA
	var prevBold bool

	_, maxRow, _, _ := w.canvas.GetContentBounds()

	for r := 0; r <= maxRow; r++ {
		row := w.canvas.Grid[r]
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

		// Reset attributes at the start of each line for clean state.
		prevFg, prevBg = color.RGBA{}, color.RGBA{}
		prevBold = false
		fmt.Fprint(w.writer, "\x1b[0m")

		for i := 0; i <= lastCharIndex; i++ {
			cell := row[i]
			if cell.Fg != prevFg || cell.Bg != prevBg || cell.Bold != prevBold {
				var params []string

				// Find the closest ANSI color index for FG and BG
				_, fgIndex := FindClosestAnsiColor(cell.Fg)
				_, bgIndex := FindClosestAnsiColor(cell.Bg)

				// Determine if the colors are in the "bright" range (8-15)
				fgIsBright := fgIndex >= 8
				bgIsBright := bgIndex >= 8

				if cell.Bold {
					params = append(params, "1")
				} else {
					params = append(params, "22") // Non-bold
				}

				if fgIsBright {
					params = append(params, fmt.Sprintf("%d", (fgIndex-8)+90))
				} else {
					params = append(params, fmt.Sprintf("%d", fgIndex+30))
				}

				if bgIsBright {
					params = append(params, fmt.Sprintf("%d", (bgIndex-8)+100))
				} else {
					params = append(params, fmt.Sprintf("%d", bgIndex+40))
				}

				fmt.Fprintf(w.writer, "\x1b[%sm", strings.Join(params, ";"))

				prevFg = cell.Fg
				prevBg = cell.Bg
				prevBold = cell.Bold
			}
			if _, err := w.writer.Write([]byte(string(cell.Char))); err != nil {
				return err
			}
		}
		fmt.Fprint(w.writer, "\x1b[0m\n")
	}

	return nil
}
