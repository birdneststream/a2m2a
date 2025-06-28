package ansi

import (
	"a2m2a/canvas"
	"bufio"
	"io"

	"golang.org/x/text/encoding/charmap"
)

// Parser holds the state for parsing an ANSI stream.
type Parser struct {
	canvas *canvas.Canvas
	reader *bufio.Reader
	// Current graphic rendition attributes
	fg   int
	bg   int
	bold bool
	ice  bool
}

// NewParser creates a new ANSI parser.
func NewParser(c *canvas.Canvas, r io.Reader) *Parser {
	// The input stream is decoded from CP437 to UTF-8.
	decodedReader := charmap.CodePage437.NewDecoder().Reader(r)
	return &Parser{
		canvas: c,
		reader: bufio.NewReader(decodedReader),
		fg:     canvas.DefaultFg,
		bg:     canvas.DefaultBg,
		bold:   canvas.DefaultBold,
		ice:    canvas.DefaultIce,
	}
}

// Parse reads the ANSI stream and updates the canvas.
func (p *Parser) Parse() error {
	for {
		r, _, err := p.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch r {
		case '\x1b': // ESC
			if err := p.handleEscape(); err != nil {
				if err == io.EOF {
					return nil
				}
				// It's common for ANSI art to have truncated escape codes.
				// We can log this as a warning instead of returning an error.
				// For now, we'll just stop parsing.
				return nil
			}
		case '\n':
			p.canvas.NewLine()
		case '\r':
			p.canvas.Cursor.Col = 0
		case '\t':
			// In the C code, tab size is configurable. We'll hardcode 8 for now.
			const tabSize = 8
			for i := 0; i < tabSize; i++ {
				p.canvas.SetCell(' ', p.fg, p.bg, p.bold, p.ice)
			}
		case '\x1a': // SAUCE terminator
			return nil
		default:
			p.canvas.SetCell(r, p.fg, p.bg, p.bold, p.ice)
		}
	}
}

func (p *Parser) handleEscape() error {
	r, _, err := p.reader.ReadRune()
	if err != nil {
		return err
	}

	if r != '[' {
		// Not a CSI sequence, ignore for now.
		return nil
	}

	// This is a Control Sequence Introducer (CSI)
	// It has the form: \x1b[<params>...<command>
	params := make([]int, 0, 16)
	currentParam := 0
	inParam := false

	for {
		r, _, err := p.reader.ReadRune()
		if err != nil {
			return err
		}

		if r >= '0' && r <= '9' {
			if !inParam {
				inParam = true
			}
			currentParam = (currentParam * 10) + int(r-'0')
			continue
		}

		// Delimiter or command
		if inParam {
			params = append(params, currentParam)
			currentParam = 0
			inParam = false
		}

		if r == ';' {
			continue
		}

		// This is a command rune
		p.executeCommand(r, params)
		return nil
	}
}

func (p *Parser) executeCommand(cmd rune, params []int) {
	getParam := func(index, defaultValue int) int {
		if index < len(params) {
			return params[index]
		}
		return defaultValue
	}

	switch cmd {
	case 'm': // Select Graphic Rendition (SGR)
		if len(params) == 0 {
			params = []int{0} // Reset
		}
		for i := 0; i < len(params); i++ {
			param := params[i]
			switch {
			case param == 0: // Reset
				p.fg = canvas.DefaultFg
				p.bg = canvas.DefaultBg
				p.bold = canvas.DefaultBold
				p.ice = canvas.DefaultIce
			case param == 1:
				p.bold = true
			case param == 5:
				p.ice = true
			case param == 22:
				p.bold = false
			case param == 25:
				p.ice = false
			case param >= 30 && param <= 37:
				p.fg = param - 30
			case param == 39:
				p.fg = canvas.DefaultFg
			case param >= 40 && param <= 47:
				p.bg = param - 40
			case param == 49:
				p.bg = canvas.DefaultBg
			case param >= 90 && param <= 97: // high intensity foreground
				p.fg = param - 90
				p.bold = true
			case param >= 100 && param <= 107: // high intensity background
				p.bg = param - 100
				p.ice = true
			}
		}
	case 'H', 'f': // Cursor Position
		row := getParam(0, 1)
		col := getParam(1, 1)
		p.canvas.MoveTo(row, col)
	case 'A': // Cursor Up
		p.canvas.MoveUp(getParam(0, 1))
	case 'B': // Cursor Down
		p.canvas.MoveDown(getParam(0, 1))
	case 'C': // Cursor Forward
		p.canvas.MoveForward(getParam(0, 1))
	case 'D': // Cursor Back
		p.canvas.MoveBackward(getParam(0, 1))
	case 'J': // Erase in Display
		// 2J clears the entire screen
		if getParam(0, 0) == 2 {
			p.canvas.ClearScreen()
		}
		// Other J commands are not implemented for simplicity
	case 'K': // Erase in Line
		// Not implemented for simplicity
	case 's': // Save Cursor Position
		p.canvas.SaveCursor()
	case 'u': // Restore Cursor Position
		p.canvas.RestoreCursor()
	}
}
