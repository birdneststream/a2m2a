package renderer

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"a2m2a/canvas"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// ansiColorPalette maps the 16 ANSI colors to RGBA values.
var ansiColorPalette = []color.RGBA{
	{0, 0, 0, 255},       // 0: Black
	{170, 0, 0, 255},     // 1: Red
	{0, 170, 0, 255},     // 2: Green
	{170, 85, 0, 255},    // 3: Yellow (Brown)
	{0, 0, 170, 255},     // 4: Blue
	{170, 0, 170, 255},   // 5: Magenta
	{0, 170, 170, 255},   // 6: Cyan
	{170, 170, 170, 255}, // 7: White (Light Gray)
	{85, 85, 85, 255},    // 8: Bright Black (Dark Gray)
	{255, 85, 85, 255},   // 9: Bright Red
	{85, 255, 85, 255},   // 10: Bright Green
	{255, 255, 85, 255},  // 11: Bright Yellow
	{85, 85, 255, 255},   // 12: Bright Blue
	{255, 85, 255, 255},  // 13: Bright Magenta
	{85, 255, 255, 255},  // 14: Bright Cyan
	{255, 255, 255, 255}, // 15: Bright White
}

const (
	// These are typical dimensions for VGA text mode fonts.
	charWidth  = 9
	charHeight = 16
	fontSize   = 16
)

// ToPNG renders a canvas to a PNG image.
func ToPNG(c *canvas.Canvas) ([]byte, error) {
	parsedFont, err := truetype.Parse(FontData)
	if err != nil {
		return nil, err
	}
	img := renderCanvasToImage(c, parsedFont, 1.0)
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ToThumbnail generates a small PNG thumbnail from the canvas.
func ToThumbnail(c *canvas.Canvas) ([]byte, error) {
	parsedFont, err := truetype.Parse(FontData)
	if err != nil {
		return nil, err
	}
	// Use a smaller scale for the thumbnail. 0.5 means half the size.
	img := renderCanvasToImage(c, parsedFont, 0.5)
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// renderCanvasToImage performs the actual drawing of the canvas to an image.
func renderCanvasToImage(c *canvas.Canvas, parsedFont *truetype.Font, scale float64) image.Image {
	// Determine the actual bounds of the art to create a tightly-cropped image.
	minRow, maxRow, minCol, maxCol := c.GetContentBounds()
	if minRow > maxRow { // Empty canvas
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	const baseFontSize = 16.0 // A standard size for getting metrics
	face := truetype.NewFace(parsedFont, &truetype.Options{
		Size:    baseFontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})

	// Correctly calculate character dimensions from font metrics.
	// The font metrics are in 26.6 fixed-point format, so we divide by 64.
	advance, _ := face.GlyphAdvance('M')
	fCharWidth := (float64(advance) / 64.0) * scale
	fCharHeight := (float64(face.Metrics().Ascent+face.Metrics().Descent) / 64.0) * scale
	baseline := int((float64(face.Metrics().Ascent) / 64.0) * scale)

	numCols := maxCol - minCol + 1
	numRows := maxRow - minRow + 1

	imgWidth := int(float64(numCols) * fCharWidth)
	imgHeight := int(float64(numRows) * fCharHeight)

	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.Black}, image.Point{}, draw.Src)

	// Create a new face for rendering with the scaled size
	scaledFace := truetype.NewFace(parsedFont, &truetype.Options{
		Size:    baseFontSize * scale,
		DPI:     72,
		Hinting: font.HintingFull,
	})

	drawer := &font.Drawer{
		Dst:  img,
		Face: scaledFace,
	}

	for r := minRow; r <= maxRow; r++ {
		for col := minCol; col <= maxCol; col++ {
			cell := c.Grid[r][col]

			// Determine foreground and background colors
			fgIndex := cell.Fg
			if cell.Bright {
				fgIndex += 8 // Bright colors are in the upper half of the palette
			} else if cell.Bold && cell.Fg == 0 {
				// Special case for bold black, which should be rendered as dark gray.
				fgIndex += 8
			}
			fgColor := ansiColorPalette[fgIndex]

			bgIndex := cell.Bg
			if cell.Ice {
				bgIndex += 8 // iCE colors use the bright palette for backgrounds
			}
			bgColor := ansiColorPalette[bgIndex]

			// Calculate the pixel boundaries for the cell.
			startX := (col - minCol) * imgWidth / numCols
			startY := (r - minRow) * imgHeight / numRows
			endX := (col - minCol + 1) * imgWidth / numCols
			endY := (r - minRow + 1) * imgHeight / numRows

			// Special handling for block-drawing characters to ensure pixel-perfect rendering.
			switch cell.Char {
			case '█': // Full block
				draw.Draw(img, image.Rect(startX, startY, endX, endY), &image.Uniform{C: fgColor}, image.Point{}, draw.Src)
				continue
			case '▀': // Upper half block
				midY := startY + (endY-startY)/2
				draw.Draw(img, image.Rect(startX, startY, endX, midY), &image.Uniform{C: fgColor}, image.Point{}, draw.Src)
				draw.Draw(img, image.Rect(startX, midY, endX, endY), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)
				continue
			case '▄': // Lower half block
				midY := startY + (endY-startY)/2
				draw.Draw(img, image.Rect(startX, startY, endX, midY), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)
				draw.Draw(img, image.Rect(startX, midY, endX, endY), &image.Uniform{C: fgColor}, image.Point{}, draw.Src)
				continue
			case '▌': // Left half block
				midX := startX + (endX-startX)/2
				draw.Draw(img, image.Rect(startX, startY, midX, endY), &image.Uniform{C: fgColor}, image.Point{}, draw.Src)
				draw.Draw(img, image.Rect(midX, startY, endX, endY), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)
				continue
			case '▐': // Right half block
				midX := startX + (endX-startX)/2
				draw.Draw(img, image.Rect(startX, startY, midX, endY), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)
				draw.Draw(img, image.Rect(midX, startY, endX, endY), &image.Uniform{C: fgColor}, image.Point{}, draw.Src)
				continue
			case '▓': // Dark shade (75%)
				draw.Draw(img, image.Rect(startX, startY, endX, endY), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)
				for y := startY; y < endY; y++ {
					for x := startX; x < endX; x++ {
						// 3/4 dot pattern
						if x%2 == 1 || y%2 == 1 {
							img.Set(x, y, fgColor)
						}
					}
				}
				continue
			case '▒': // Medium shade (50%)
				draw.Draw(img, image.Rect(startX, startY, endX, endY), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)
				for y := startY; y < endY; y++ {
					for x := startX; x < endX; x++ {
						// 50% checkerboard pattern
						if (x+y)%2 == 0 {
							img.Set(x, y, fgColor)
						}
					}
				}
				continue
			case '░': // Light shade (25%)
				draw.Draw(img, image.Rect(startX, startY, endX, endY), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)
				for y := startY; y < endY; y++ {
					for x := startX; x < endX; x++ {
						// 1/4 dot pattern
						if x%2 == 0 && y%2 == 0 {
							img.Set(x, y, fgColor)
						}
					}
				}
				continue
			}

			// Default rendering for all other characters.
			bgRect := image.Rect(startX, startY, endX, endY)
			draw.Draw(img, bgRect, &image.Uniform{C: bgColor}, image.Point{}, draw.Src)

			// Draw foreground character, skipping spaces.
			if cell.Char != ' ' {
				drawer.Src = &image.Uniform{C: fgColor}
				drawer.Dot = fixed.Point26_6{
					X: fixed.I(startX),
					Y: fixed.I(startY + baseline),
				}
				drawer.DrawString(string(cell.Char))
			}
		}
	}

	return img
}
