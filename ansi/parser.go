package ansi

import (
	"a2m2a/canvas"
	"bufio"
	"io"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

// Parser holds the state for parsing an ANSI stream.
type Parser struct {
	canvas      *canvas.Canvas
	reader      *bufio.Reader
	savedCursor canvas.Point // For DECSC and DECRC
	// Current graphic rendition attributes
	fg     int
	bg     int
	bold   bool
	bright bool
	ice    bool
}

// NewParser creates a new ANSI parser.
func NewParser(c *canvas.Canvas, r io.Reader, dataSize int64) *Parser {
	var limitedReader io.Reader
	if dataSize > 0 {
		limitedReader = io.LimitReader(r, dataSize)
	} else {
		limitedReader = r
	}
	// The input stream is decoded from CP437 to UTF-8.
	decodedReader := charmap.CodePage437.NewDecoder().Reader(limitedReader)
	return &Parser{
		canvas:      c,
		reader:      bufio.NewReader(decodedReader),
		savedCursor: canvas.Point{Row: 0, Col: 0},
		fg:          canvas.DefaultFg,
		bg:          canvas.DefaultBg,
		bold:        canvas.DefaultBold,
		bright:      false,
		ice:         canvas.DefaultIce,
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
		case '\x1b': // Escape character
			if err := p.handleEscape(); err != nil {
				return err
			}
		case '\n':
			p.canvas.NewLine()
		case '\r':
			p.canvas.Cursor.Col = 0
		case '\t':
			// In the C code, tab size is configurable. We'll hardcode 8 for now.
			const tabSize = 8
			for i := 0; i < tabSize; i++ {
				p.canvas.SetCell(' ', p.fg, p.bg, p.bold, p.bright, p.ice)
			}
		case '\x1a': // SAUCE separator. Should be handled by LimitReader now, but we keep this for safety.
			return nil
		default:
			p.canvas.SetCell(r, p.fg, p.bg, p.bold, p.bright, p.ice)
		}
	}
}

func (p *Parser) handleEscape() error {
	r, _, err := p.reader.ReadRune()
	if err != nil {
		return err
	}

	if r == '[' { // This is a Control Sequence Introducer (CSI)
		return p.handleCSI()
	}

	// Other escape sequences are not supported for now.
	return nil
}

func (p *Parser) handleCSI() error {
	var params []int
	var currentParam strings.Builder
	var cmd rune

	for {
		r, _, err := p.reader.ReadRune()
		if err != nil {
			return err
		}

		if r >= '0' && r <= '9' {
			currentParam.WriteRune(r)
		} else {
			if currentParam.Len() > 0 {
				val, _ := strconv.Atoi(currentParam.String())
				params = append(params, val)
				currentParam.Reset()
			}
			if r == ';' {
				if len(params) == 0 {
					params = append(params, 0)
				}
				continue
			}
			cmd = r
			break
		}
	}

	p.executeCommand(cmd, params)
	return nil
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
			params = []int{0} // Treat `[m` as `[0m`
		}
		for _, param := range params {
			switch {
			case param == 0: // Reset
				p.fg = canvas.DefaultFg
				p.bg = canvas.DefaultBg
				p.bold = canvas.DefaultBold
				p.bright = false
				p.ice = canvas.DefaultIce
			case param == 1:
				p.bold = true
			case param == 5:
				p.ice = true
			case param == 22:
				p.bold = false
				p.bright = false
			case param == 25:
				p.ice = false
			case param >= 30 && param <= 37:
				p.fg = param - 30
				p.bright = false // Standard colors are not bright
			case param == 39:
				p.fg = canvas.DefaultFg
			case param >= 40 && param <= 47:
				p.bg = param - 40
			case param == 49:
				p.bg = canvas.DefaultBg
			case param >= 90 && param <= 97: // high intensity foreground
				p.fg = param - 90
				p.bright = true
			case param >= 100 && param <= 107: // high intensity background
				p.bg = param - 100
				p.ice = true
			}
		}
	case 'H', 'f': // Cursor Position
		row := getParam(0, 1)
		col := getParam(1, 1)
		p.canvas.SetCursor(row-1, col-1) // ANSI is 1-based
	case 'A': // Cursor Up
		p.canvas.MoveUp(getParam(0, 1))
	case 'B': // Cursor Down
		p.canvas.MoveDown(getParam(0, 1))
	case 'C': // Cursor Forward
		p.canvas.MoveForward(getParam(0, 1))
	case 'D': // Cursor Backward
		p.canvas.MoveBackward(getParam(0, 1))
	case 'J': // Erase in Display
		mode := getParam(0, 0)
		switch mode {
		case 2: // Erase entire screen and move cursor to home
			p.canvas.Clear(canvas.Cell{
				Char:   ' ',
				Fg:     p.fg,
				Bg:     p.bg,
				Bold:   p.bold,
				Bright: p.bright,
				Ice:    p.ice,
			})
		}
	case 's': // Save Cursor Position (SCOSC/DECSC)
		p.savedCursor = p.canvas.Cursor
	case 'u': // Restore Cursor Position (SCORC/DECRC)
		p.canvas.Cursor = p.savedCursor
	}
}
