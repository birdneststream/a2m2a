package mirc

import (
	"a2m2a/canvas"
	"bufio"
	"io"
	"strconv"
	"unicode"
)

type ansiAttribute struct {
	index int
	bold  bool
}

var mircToAnsi map[int]ansiAttribute

func init() {
	mircToAnsi = make(map[int]ansiAttribute)
	// Populate from the writer's color maps
	for i, mircColor := range color {
		mircToAnsi[mircColor] = ansiAttribute{index: i, bold: false}
	}
	for i, mircColor := range colorBold {
		// If a color is in both maps (like black/white), bold takes precedence
		// as it's the more specific case.
		mircToAnsi[mircColor] = ansiAttribute{index: i, bold: true}
	}
}

// Parser holds the state for parsing a mIRC stream.
type Parser struct {
	canvas *canvas.Canvas
	reader *bufio.Reader
	// Current graphic rendition attributes
	fg   int
	bg   int
	bold bool
	ice  bool
}

// NewParser creates a new mIRC parser.
func NewParser(c *canvas.Canvas, r io.Reader) *Parser {
	return &Parser{
		canvas: c,
		reader: bufio.NewReader(r),
		fg:     canvas.DefaultFg,
		bg:     canvas.DefaultBg,
		bold:   canvas.DefaultBold,
		ice:    canvas.DefaultIce,
	}
}

// Parse reads the mIRC stream and updates the canvas.
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
		case '\x03': // mIRC color code
			if err := p.handleColorCode(); err != nil {
				// Stop parsing on any error from color handler, including EOF
				return nil
			}
		case '\n':
			p.canvas.NewLine()
		case '\r':
			p.canvas.Cursor.Col = 0
		case '\x02': // Bold toggle
			p.bold = !p.bold
		case '\x1d': // Italic toggle (unsupported)
			continue
		case '\x1f': // Underline toggle (unsupported)
			continue
		default:
			p.canvas.SetCell(r, p.fg, p.bg, p.bold, false, p.ice)
		}
	}
}

func (p *Parser) handleColorCode() error {
	// Color code format: \x03<FG>[<,BG>]
	// FG and BG are 1 or 2 digits.

	// Read foreground
	fgStr, err := p.readColorDigits()
	if err != nil {
		// This can happen if \x03 is at the end of the file.
		// It's not a fatal error, just stop parsing.
		if err == io.EOF {
			return nil
		}
		// If there are no digits after \x03, it's a reset code.
		p.fg, p.bg = canvas.DefaultFg, canvas.DefaultBg
		p.bold, p.ice = canvas.DefaultBold, canvas.DefaultIce
		return nil
	}

	fgColor, _ := strconv.Atoi(fgStr)
	if attr, ok := mircToAnsi[fgColor]; ok {
		p.fg = attr.index
		p.bold = attr.bold
	}

	// Check for optional background
	nextRune, _, err := p.reader.ReadRune()
	if err != nil {
		return err // EOF during BG check
	}

	if nextRune != ',' {
		// No BG color, unread the rune and we're done.
		p.reader.UnreadRune()
		return nil
	}

	// Read background
	bgStr, err := p.readColorDigits()
	if err != nil {
		// EOF after comma
		return nil
	}

	bgColor, _ := strconv.Atoi(bgStr)
	if attr, ok := mircToAnsi[bgColor]; ok {
		p.bg = attr.index
		// The "iCE" convention uses bold/bright colors for backgrounds
		p.ice = attr.bold
	}

	return nil
}

// readColorDigits reads 1 or 2 digits from the reader.
func (p *Parser) readColorDigits() (string, error) {
	var digits []rune
	// Read first digit
	r, _, err := p.reader.ReadRune()
	if err != nil {
		return "", err
	}
	if !unicode.IsDigit(r) {
		p.reader.UnreadRune()
		return "", io.EOF // Not a digit, treat as end
	}
	digits = append(digits, r)

	// Read optional second digit
	r, _, err = p.reader.ReadRune()
	if err != nil {
		// EOF is fine after one digit
		return string(digits), nil
	}
	if !unicode.IsDigit(r) {
		p.reader.UnreadRune()
		return string(digits), nil
	}
	digits = append(digits, r)

	return string(digits), nil
}
