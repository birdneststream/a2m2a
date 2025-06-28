package canvas

import "image/color"

var (
	// These are standard VGA colors, which are the basis for ANSI color.
	// We'll use them for defaults.
	DefaultFg = color.RGBA{R: 0xAA, G: 0xAA, B: 0xAA, A: 0xFF} // Light Grey (ANSI 7)
	DefaultBg = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF} // Black (ANSI 0)
)

const (
	DefaultBold  = false
	DefaultIce   = false
	DefaultWidth = 80
)

// Point represents a coordinate on the canvas.
type Point struct {
	Row int
	Col int
}

// Cell represents a single character cell on the canvas.
type Cell struct {
	Char   rune
	Fg     color.RGBA
	Bg     color.RGBA
	Bold   bool // For font weight (SGR 1)
	Bright bool // For high-intensity colors (SGR 90-97)
	Ice    bool // For high-intensity backgrounds (iCE Color / SGR 5)
}

// Canvas represents the grid of characters.
type Canvas struct {
	Grid   [][]Cell
	Cursor Point
	width  int
}

// NewCanvas creates a new canvas of a given width.
func NewCanvas(width int) *Canvas {
	c := &Canvas{
		width: width,
	}
	c.addRow()
	return c
}

// SetCursor moves the cursor to an absolute position.
func (c *Canvas) SetCursor(row, col int) {
	if row < 0 {
		row = 0
	}
	if col < 0 {
		col = 0
	}
	// Ensure enough rows exist
	for row >= len(c.Grid) {
		c.addRow()
	}
	c.Cursor.Row = row
	c.Cursor.Col = col
}

// MoveUp moves the cursor up by n rows.
func (c *Canvas) MoveUp(n int) {
	c.Cursor.Row -= n
	if c.Cursor.Row < 0 {
		c.Cursor.Row = 0
	}
}

// MoveDown moves the cursor down by n rows.
func (c *Canvas) MoveDown(n int) {
	c.Cursor.Row += n
	for c.Cursor.Row >= len(c.Grid) {
		c.addRow()
	}
}

// MoveForward moves the cursor forward by n columns.
func (c *Canvas) MoveForward(n int) {
	c.Cursor.Col += n
	if c.Cursor.Col >= c.width {
		c.Cursor.Col = c.width - 1
	}
}

// MoveBackward moves the cursor backward by n columns.
func (c *Canvas) MoveBackward(n int) {
	c.Cursor.Col -= n
	if c.Cursor.Col < 0 {
		c.Cursor.Col = 0
	}
}

// addRow adds a new row to the canvas grid.
func (c *Canvas) addRow() {
	newRow := make([]Cell, c.width)
	for i := range newRow {
		newRow[i] = Cell{
			Char:   ' ',
			Fg:     DefaultFg,
			Bg:     DefaultBg,
			Bold:   DefaultBold,
			Bright: false, // Default for Bright is false
			Ice:    DefaultIce,
		}
	}
	c.Grid = append(c.Grid, newRow)
}

// NewLine moves the cursor to the beginning of the next line.
func (c *Canvas) NewLine() {
	c.Cursor.Col = 0
	c.Cursor.Row++
	if c.Cursor.Row >= len(c.Grid) {
		c.addRow()
	}
}

// SetCell places a character at the current cursor position and advances the cursor.
func (c *Canvas) SetCell(char rune, fg, bg color.RGBA, bold, bright, ice bool) {
	if c.Cursor.Row >= len(c.Grid) {
		c.addRow()
	}

	// This check prevents writing past the last column, which can happen
	// with some ANSI files that don't respect their own width.
	if c.Cursor.Col >= c.width {
		c.NewLine()
	}

	c.Grid[c.Cursor.Row][c.Cursor.Col] = Cell{
		Char:   char,
		Fg:     fg,
		Bg:     bg,
		Bold:   bold,
		Bright: bright,
		Ice:    ice,
	}

	c.Cursor.Col++
	if c.Cursor.Col >= c.width {
		c.Cursor.Col = 0
		c.Cursor.Row++
	}
}

// Clear sets the entire canvas to a specific cell state and resets the cursor.
func (c *Canvas) Clear(cell Cell) {
	for r := range c.Grid {
		for col := range c.Grid[r] {
			c.Grid[r][col] = cell
			c.Grid[r][col].Char = ' ' // Ensure the character is a space
		}
	}
	c.Cursor.Row = 0
	c.Cursor.Col = 0
}

// GetContentBounds finds the minimal bounding box of content on the canvas.
func (c *Canvas) GetContentBounds() (minRow, maxRow, minCol, maxCol int) {
	minRow, maxRow, minCol, maxCol = len(c.Grid), -1, c.width, -1

	for r, row := range c.Grid {
		for col, cell := range row {
			// A cell has content if it's not a space with a default background.
			if cell.Char != ' ' || cell.Bg != DefaultBg {
				if r < minRow {
					minRow = r
				}
				if r > maxRow {
					maxRow = r
				}
				if col < minCol {
					minCol = col
				}
				if col > maxCol {
					maxCol = col
				}
			}
		}
	}
	// If no content was found, return zero values.
	if maxRow == -1 {
		return 0, 0, 0, 0
	}
	return minRow, maxRow, minCol, maxCol
}
