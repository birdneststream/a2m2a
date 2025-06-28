package mirc

import (
	"a2m2a/ansi"
	"a2m2a/canvas"
	"bufio"
	clr "image/color"
	"io"
	"strconv"
	"unicode"
)

// Parser holds the state for parsing a mIRC stream.
type Parser struct {
	canvas  *canvas.Canvas
	reader  *bufio.Reader
	force16 bool
	// Current graphic rendition attributes
	fg   clr.RGBA
	bg   clr.RGBA
	bold bool
	ice  bool
}

// NewParser creates a new mIRC parser.
func NewParser(c *canvas.Canvas, r io.Reader, force16 bool) *Parser {
	return &Parser{
		canvas:  c,
		reader:  bufio.NewReader(r),
		force16: force16,
		fg:      canvas.DefaultFg,
		bg:      canvas.DefaultBg,
		bold:    canvas.DefaultBold,
		ice:     canvas.DefaultIce,
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
			// If the cursor is not at the start of a line, we need to add a newline.
			// If it *is* at the start, it means the canvas auto-wrapped for us,
			// and we should not add a second newline.
			if p.canvas.Cursor.Col != 0 {
				p.canvas.NewLine()
			}
		case '\r':
			// Treat '\r' as a full newline. If a '\n' follows, the logic
			// in the '\n' case will correctly prevent a double newline because
			// the cursor column will already be 0.
			p.canvas.NewLine()
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
		p.bold, p.ice = false, false
		return nil
	}

	fgColorIdx, _ := strconv.Atoi(fgStr)
	if fgColorIdx >= 0 && fgColorIdx < len(MircPalette99) {
		color := MircPalette99[fgColorIdx]
		if p.force16 {
			color, _ = ansi.FindClosestAnsiColor(color)
		}
		p.fg = color
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

	bgColorIdx, _ := strconv.Atoi(bgStr)
	if bgColorIdx >= 0 && bgColorIdx < len(MircPalette99) {
		color := MircPalette99[bgColorIdx]
		if p.force16 {
			color, _ = ansi.FindClosestAnsiColor(color)
		}
		p.bg = color
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
