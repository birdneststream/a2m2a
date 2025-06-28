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
	var writer io.WriteCloser = os.Stdout // Default to stdout

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

	// --- Auto-Detect Format ---
	// We need to peek at the first part of the file without consuming the reader.
	buf := new(bytes.Buffer)
	// TeeReader lets us read while also writing to a buffer
	tee := io.TeeReader(reader, buf)
	format := detectFormat(tee)
	// Now, chain the buffer back to the start of the reader
	reader = io.MultiReader(buf, reader)

	// --- Output Handling ---
	if outPath != "" {
		// Ensure the output file has the correct extension for image generation.
		if (png || thumb > 0) && !strings.HasSuffix(outPath, ".png") {
			outPath += ".png"
		}
		file, err := os.Create(outPath)
		if err != nil {
			log.Fatalf("Error creating output file: %v", err)
		}
		defer file.Close()
		writer = file
	}

	// --- Conversion Logic ---
	c := canvas.NewCanvas(width)
	var outputFormat string

	switch format {
	case "ansi":
		p := ansi.NewParser(c, reader)
		if err := p.Parse(); err != nil {
			log.Fatalf("Error parsing ANSI: %v", err)
		}
		if !png && thumb == 0 {
			outputFormat = "mirc"
		}
	case "mirc":
		p := mirc.NewParser(c, reader)
		if err := p.Parse(); err != nil {
			log.Fatalf("Error parsing mIRC: %v", err)
		}
		// If we are generating an image, we don't need to know the text output format.
		if !png && thumb == 0 {
			outputFormat = "ansi"
		}
	default:
		log.Fatalf("Could not detect file format. Please specify manually.")
	}

	// --- Output Generation ---
	if png {
		if err := renderer.ToPNG(c, writer); err != nil {
			log.Fatalf("Error generating PNG: %v", err)
		}
	} else if thumb > 0 {
		if err := renderer.ToThumbnail(c, writer, thumb); err != nil {
			log.Fatalf("Error generating thumbnail: %v", err)
		}
	} else {
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
	}

	fmt.Println("Conversion complete.")
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
