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

	face := truetype.NewFace(parsedFont, &truetype.Options{
		Size:    16, // Use a base size
		DPI:     72,
		Hinting: font.HintingFull,
	})

	// Calculate character dimensions from font metrics.
	bounds, advance, _ := face.GlyphBounds('M') // Use 'M' as a representative character
	charWidth := int(float64(advance) * scale)
	charHeight := int(float64(bounds.Max.Y-bounds.Min.Y) * scale)

	imgWidth := (maxCol - minCol + 1) * charWidth
	imgHeight := (maxRow - minRow + 1) * charHeight

	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.Black}, image.Point{}, draw.Src)

	// Create a new face for rendering with the scaled size
	scaledFace := truetype.NewFace(parsedFont, &truetype.Options{
		Size:    16 * scale,
		DPI:     72,
		Hinting: font.HintingFull,
	})

	for r := minRow; r <= maxRow; r++ {
		for col := minCol; col <= maxCol; col++ {
			cell := c.Grid[r][col]

			// Determine foreground and background colors
			fgIndex := cell.Fg
			if cell.Bright {
				fgIndex += 8 // Bright colors are in the upper half of the palette
			}
			fgColor := ansiColorPalette[fgIndex]

			bgIndex := cell.Bg
			if cell.Ice {
				bgIndex += 8 // iCE colors use the bright palette for backgrounds
			}
			bgColor := ansiColorPalette[bgIndex]

			// Draw background
			bgRect := image.Rect(
				(col-minCol)*charWidth,
				(r-minRow)*charHeight,
				(col-minCol+1)*charWidth,
				(r-minRow+1)*charHeight,
			)
			draw.Draw(img, bgRect, &image.Uniform{C: bgColor}, image.Point{}, draw.Src)

			// Draw foreground character
			if cell.Char != ' ' {
				drawer := &font.Drawer{
					Dst:  img,
					Src:  &image.Uniform{C: fgColor},
					Face: scaledFace,
					Dot:  fixed.Point26_6{X: fixed.I((col - minCol) * charWidth), Y: fixed.I((r-minRow+1)*charHeight - int(float64(scaledFace.Metrics().Descent)))},
				}
				drawer.DrawString(string(cell.Char))
			}
		}
	}

	return img
}
