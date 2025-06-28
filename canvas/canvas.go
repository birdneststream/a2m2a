package canvas

const (
	DefaultFg    = 7
	DefaultBg    = 0
	DefaultBold  = false
	DefaultIce   = false
	DefaultWidth = 80
)

// Cell represents a single character cell on the canvas.
type Cell struct {
	Char rune
	Fg   int
	Bg   int
	Bold bool
	Ice  bool
}

// Canvas represents the grid of characters.
type Canvas struct {
	Grid        [][]Cell
	width       int
	Cursor      Cursor
	savedCursor Cursor
}

// Cursor represents the position of the cursor on the canvas.
type Cursor struct {
	Row int
	Col int
}

// NewCanvas creates a new canvas with the given width.
func NewCanvas(width int) *Canvas {
	if width <= 0 {
		width = DefaultWidth
	}
	c := &Canvas{
		width: width,
		Grid:  make([][]Cell, 0),
	}
	c.addRow()
	return c
}

// addRow adds a new row to the canvas.
func (c *Canvas) addRow() {
	newRow := make([]Cell, c.width)
	for i := range newRow {
		newRow[i] = Cell{
			Char: ' ',
			Fg:   DefaultFg,
			Bg:   DefaultBg,
			Bold: DefaultBold,
			Ice:  DefaultIce,
		}
	}
	c.Grid = append(c.Grid, newRow)
}

// SetCell places a character at the current cursor position and advances the cursor.
func (c *Canvas) SetCell(char rune, fg, bg int, bold, ice bool) {
	if c.Cursor.Row >= len(c.Grid) {
		c.addRow()
	}
	if c.Cursor.Col >= c.width {
		c.NewLine()
	}

	c.Grid[c.Cursor.Row][c.Cursor.Col] = Cell{
		Char: char,
		Fg:   fg,
		Bg:   bg,
		Bold: bold,
		Ice:  ice,
	}

	c.Cursor.Col++
	if c.Cursor.Col >= c.width {
		c.NewLine()
	}
}

// NewLine moves the cursor to the beginning of the next line.
func (c *Canvas) NewLine() {
	c.Cursor.Col = 0
	c.Cursor.Row++
	if c.Cursor.Row >= len(c.Grid) {
		c.addRow()
	}
}

// MoveTo moves the cursor to a specific row and column.
func (c *Canvas) MoveTo(row, col int) {
	// ANSI is 1-based, converting to 0-based.
	r := row - 1
	if r < 0 {
		r = 0
	}
	for r >= len(c.Grid) {
		c.addRow()
	}
	c.Cursor.Row = r

	c.Cursor.Col = col - 1
	if c.Cursor.Col < 0 {
		c.Cursor.Col = 0
	}
	if c.Cursor.Col >= c.width {
		c.Cursor.Col = c.width - 1
	}
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
		// In ANSI, moving past the last column can have different behaviors.
		// Here, we'll just clamp it to the last column.
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

// SaveCursor saves the current cursor position.
func (c *Canvas) SaveCursor() {
	c.savedCursor = c.Cursor
}

// RestoreCursor restores the saved cursor position.
func (c *Canvas) RestoreCursor() {
	c.Cursor = c.savedCursor
}

// ClearScreen clears the entire canvas.
func (c *Canvas) ClearScreen() {
	c.Cursor.Row = 0
	c.Cursor.Col = 0
	c.Grid = make([][]Cell, 0)
	c.addRow()
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

	// If canvas is empty, return zero values.
	if maxRow == -1 {
		return 0, 0, 0, 0
	}

	return minRow, maxRow, minCol, maxCol
}
