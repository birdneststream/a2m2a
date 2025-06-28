package main

import (
	"a2m2a/ansi"
	"a2m2a/canvas"
	"a2m2a/mirc"
	"a2m2a/renderer"
	"a2m2a/sauce"
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
	force16 bool
)

func init() {
	flag.StringVar(&inPath, "i", "", "Input file path (default: stdin)")
	flag.StringVar(&inPath, "in", "", "Input file path (default: stdin)")
	flag.StringVar(&outPath, "o", "", "Output file path (default: stdout)")
	flag.StringVar(&outPath, "out", "", "Output file path (default: stdout)")
	flag.IntVar(&width, "w", 80, "Width of the canvas")
	flag.BoolVar(&png, "png", false, "Generate a PNG image")
	flag.UintVar(&thumb, "thumb", 0, "Generate a thumbnail PNG of the specified width (e.g., --thumb 320)")
	flag.BoolVar(&force16, "16", false, "Force 16-color output for all formats.")
}

func main() {
	flag.Parse()

	var reader io.Reader
	var file *os.File
	var err error
	var sauceRecord *sauce.Record

	// --- Input Handling ---
	if inPath != "" {
		file, err = os.Open(inPath)
		if err != nil {
			log.Fatalf("Error opening input file: %v", err)
		}
		defer file.Close()
		reader = file

		// Try to get a SAUCE record.
		sauceRecord, _ = sauce.Get(file)
		// Reset the file reader to the beginning after SAUCE check.
		_, _ = file.Seek(0, io.SeekStart)

	} else {
		// Read from stdin if no input file is specified
		stdinData, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("Error reading from stdin: %v", err)
		}
		reader = bytes.NewReader(stdinData)
		// Note: SAUCE parsing from stdin is not supported.
	}

	// Override canvas width if specified in SAUCE record and not by user flag.
	usedWidth := width
	var dataSize int64
	if sauceRecord != nil {
		// The `width` flag has a default value, so we need to check if it was explicitly set.
		widthFlagSet := false
		flag.Visit(func(f *flag.Flag) {
			if f.Name == "w" {
				widthFlagSet = true
			}
		})
		if sauceRecord.TInfo1 > 0 && !widthFlagSet {
			usedWidth = int(sauceRecord.TInfo1)
		}
		dataSize = int64(sauceRecord.FileSize)
	}

	// --- Auto-Detect Format & Parse to Canvas ---
	buf := new(bytes.Buffer)
	tee := io.TeeReader(reader, buf)
	format := detectFormat(tee)
	reader = io.MultiReader(buf, reader)

	c := canvas.NewCanvas(usedWidth)
	var outputFormat string

	switch format {
	case "ansi":
		p := ansi.NewParser(c, reader, dataSize)
		if err := p.Parse(); err != nil {
			log.Fatalf("Error parsing ANSI: %v", err)
		}
		outputFormat = "mirc"
	case "mirc":
		// mIRC files don't have SAUCE records, so dataSize will be 0.
		p := mirc.NewParser(c, reader, force16)
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
		if !strings.HasSuffix(outPath, ".png") && !shouldGenerateThumb {
			// If we're only generating a text file, but the output has a .png extension,
			// it implies image generation. If neither --png nor --thumb is specified,
			// we add the extension. If they are, we assume the user's intent is clear.
			outPath += ".png"
		}

		// Generate the main PNG if required
		if png || !shouldGenerateThumb { // Generate main PNG if --png is set or if it's the default action
			pngData, err := renderer.ToPNG(c)
			if err != nil {
				log.Fatalf("Error generating PNG: %v", err)
			}
			if err := os.WriteFile(outPath, pngData, 0644); err != nil {
				log.Fatalf("Error writing PNG file: %v", err)
			}
			fmt.Printf("Generated PNG: %s\n", outPath)
		}

		if shouldGenerateThumb {
			thumbPath := constructThumbPath(outPath)
			thumbData, err := renderer.ToThumbnail(c)
			if err != nil {
				log.Fatalf("Error generating thumbnail: %v", err)
			}
			if err := os.WriteFile(thumbPath, thumbData, 0644); err != nil {
				log.Fatalf("Error writing thumbnail file: %v", err)
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
