package main

import (
	"a2m2a/ansi"
	"a2m2a/canvas"
	"a2m2a/mirc"
	"a2m2a/renderer"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// The new CLI flags
var (
	inPath  string
	outPath string
	width   int
	png     bool
	thumb   uint
)

func init() {
	flag.StringVar(&inPath, "i", "", "Input file path (default: stdin)")
	flag.StringVar(&inPath, "in", "", "Input file path (default: stdin)")
	flag.StringVar(&outPath, "o", "", "Output file path (default: stdout)")
	flag.StringVar(&outPath, "out", "", "Output file path (default: stdout)")
	flag.IntVar(&width, "w", 80, "Width of the canvas")
	flag.BoolVar(&png, "png", false, "Generate a PNG image")
	flag.UintVar(&thumb, "thumb", 0, "Generate a thumbnail PNG of the specified width (e.g., --thumb 320)")
}

func main() {
	flag.Parse()

	var reader io.Reader
	// --- Input Handling ---
	if inPath != "" {
		file, err := os.Open(inPath)
		if err != nil {
			log.Fatalf("Error opening input file: %v", err)
		}
		defer file.Close()
		reader = file
	} else {
		// Read from stdin if no input file is specified
		stdinData, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("Error reading from stdin: %v", err)
		}
		reader = bytes.NewReader(stdinData)
	}

	// --- Auto-Detect Format & Parse to Canvas ---
	buf := new(bytes.Buffer)
	tee := io.TeeReader(reader, buf)
	format := detectFormat(tee)
	reader = io.MultiReader(buf, reader)

	c := canvas.NewCanvas(width)
	var outputFormat string

	switch format {
	case "ansi":
		p := ansi.NewParser(c, reader)
		if err := p.Parse(); err != nil {
			log.Fatalf("Error parsing ANSI: %v", err)
		}
		outputFormat = "mirc"
	case "mirc":
		p := mirc.NewParser(c, reader)
		if err := p.Parse(); err != nil {
			log.Fatalf("Error parsing mIRC: %v", err)
		}
		outputFormat = "ansi"
	default:
		log.Fatalf("Could not detect file format. Please specify manually.")
	}

	// --- Output Generation ---
	shouldGeneratePng := png || thumb > 0 || (outPath != "" && strings.HasSuffix(outPath, ".png"))
	shouldGenerateThumb := thumb > 0

	if shouldGeneratePng {
		if outPath == "" {
			log.Fatalf("An output file path must be specified with -o or --out for image generation.")
		}
		// Ensure the output file has the correct extension for image generation.
		if !strings.HasSuffix(outPath, ".png") {
			outPath += ".png"
		}

		// Create the main PNG file writer
		pngFile, err := os.Create(outPath)
		if err != nil {
			log.Fatalf("Error creating output file: %v", err)
		}
		defer pngFile.Close()

		if err := renderer.ToPNG(c, pngFile); err != nil {
			log.Fatalf("Error generating PNG: %v", err)
		}
		fmt.Printf("Generated PNG: %s\n", outPath)

		if shouldGenerateThumb {
			thumbPath := constructThumbPath(outPath)
			thumbFile, err := os.Create(thumbPath)
			if err != nil {
				log.Fatalf("Error creating thumbnail file: %v", err)
			}
			defer thumbFile.Close()
			if err := renderer.ToThumbnail(c, thumbFile, thumb); err != nil {
				log.Fatalf("Error generating thumbnail: %v", err)
			}
			fmt.Printf("Generated Thumbnail: %s\n", thumbPath)
		}
	} else {
		// Text output logic
		var writer io.WriteCloser = os.Stdout
		if outPath != "" {
			file, err := os.Create(outPath)
			if err != nil {
				log.Fatalf("Error creating output file: %v", err)
			}
			defer file.Close()
			writer = file
		}

		switch outputFormat {
		case "ansi":
			w := ansi.NewWriter(c, writer)
			if err := w.Write(); err != nil {
				log.Fatalf("Error writing ANSI: %v", err)
			}
		case "mirc":
			w := mirc.NewWriter(c, writer)
			if err := w.Write(); err != nil {
				log.Fatalf("Error writing mIRC: %v", err)
			}
		}
		if outPath != "" {
			fmt.Printf("Generated Text File: %s\n", outPath)
		}
	}
}

// constructThumbPath creates a thumbnail filename from an original path.
// e.g., "art.png" becomes "art_thumb.png"
func constructThumbPath(originalPath string) string {
	ext := ".png"
	base := strings.TrimSuffix(originalPath, ext)
	return base + "_thumb" + ext
}

// detectFormat inspects the start of a reader to determine if it's ANSI or mIRC.
func detectFormat(r io.Reader) string {
	// Read a small chunk of the file to check for signatures.
	chunk := make([]byte, 4096)
	n, err := r.Read(chunk)
	if err != nil && err != io.EOF {
		return "unknown"
	}
	chunk = chunk[:n]

	// Look for ANSI escape code signature: \x1b[
	if bytes.Contains(chunk, []byte{0x1b, '['}) {
		return "ansi"
	}
	// Look for mIRC color code signature: \x03
	if bytes.Contains(chunk, []byte{0x03}) {
		return "mirc"
	}

	return "unknown"
}
