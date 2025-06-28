package renderer

import (
	"a2m2a/canvas"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"

	"github.com/golang/freetype/truetype"
	"github.com/nfnt/resize"
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
func ToPNG(c *canvas.Canvas, w io.Writer) error {
	img, err := renderCanvasToImage(c)
	if err != nil {
		return err
	}
	return png.Encode(w, img)
}

// ToThumbnail generates a thumbnail PNG image.
func ToThumbnail(c *canvas.Canvas, w io.Writer, width uint) error {
	img, err := renderCanvasToImage(c)
	if err != nil {
		return err
	}
	thumb := resize.Resize(width, 0, img, resize.Lanczos3)
	return png.Encode(w, thumb)
}

// renderCanvasToImage performs the actual drawing of the canvas to an image.
func renderCanvasToImage(c *canvas.Canvas) (image.Image, error) {
	parsedFont, err := truetype.Parse(FontData)
	if err != nil {
		return nil, err
	}
	face := truetype.NewFace(parsedFont, &truetype.Options{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})

	// Determine the actual bounds of the art to create a tightly-cropped image.
	minRow, maxRow, minCol, maxCol := c.GetContentBounds()
	imgWidth := (maxCol - minCol + 1) * charWidth
	imgHeight := (maxRow - minRow + 1) * charHeight

	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.Black}, image.Point{}, draw.Src)

	drawer := &font.Drawer{
		Dst:  img,
		Face: face,
	}

	for r := minRow; r <= maxRow; r++ {
		for col := minCol; col <= maxCol; col++ {
			cell := c.Grid[r][col]
			bgColor := ansiColorPalette[cell.Bg]
			fgColor := ansiColorPalette[cell.Fg]

			// Draw background
			bgRect := image.Rect(
				(col-minCol)*charWidth,
				(r-minRow)*charHeight,
				(col-minCol+1)*charWidth,
				(r-minRow+1)*charHeight,
			)
			draw.Draw(img, bgRect, &image.Uniform{C: bgColor}, image.Point{}, draw.Src)

			// Draw foreground character
			drawer.Src = &image.Uniform{C: fgColor}
			drawer.Dot = fixed.Point26_6{
				X: fixed.I((col - minCol) * charWidth),
				Y: fixed.I((r-minRow+1)*charHeight - 4), // Baseline adjustment
			}
			drawer.DrawString(string(cell.Char))
		}
	}

	return img, nil
}
