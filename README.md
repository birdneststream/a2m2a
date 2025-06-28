# a2m2a - ANSI / mIRC Art Converter

`a2m2a` (ANSI to mIRC Art to ANSI) is a versatile command-line tool written in Go for converting legacy text-based art formats and rendering them as images.

This project was entirely 'vibe-coded' into existence by an LLM in under an hour, using code from the original C utility [`a2m` by **t4t3r**](https://github.com/tat3r/a2m). It modernizes the original concept with features like automatic format detection and image rendering capabilities.

## Features

-   **Two-Way Text Conversion:**
    -   Convert ANSI art (`.ans`) to mIRC art (`.mrc`).
    -   Convert mIRC art (`.mrc`) to ANSI art (`.ans`).
-   **PNG Image Generation:**
    -   Render either ANSI or mIRC art directly to a PNG file.
    -   Creates tightly-cropped images based on the artwork's content.
-   **99-Color mIRC Support:** Accurately renders mIRC art using the full, non-standard 99-color palette, ensuring PNG outputs are true to the original.
-   **16-Color Quantization:** Can force any input into the standard 16-color ANSI palette, ensuring compatibility for text-based outputs.
-   **Thumbnail Generation:**
    -   Create a smaller thumbnail of the artwork with a user-specified width.
-   **Automatic Format Detection:** No need to specify the input format; the tool inspects the file and determines the correct conversion path automatically.
-   **CP437 Support:** Correctly handles the CP437 character set, ensuring block graphics and special symbols from classic DOS ANSI art are preserved.
-   **Flexible I/O:** Reads from and writes to files or standard input/output, allowing it to be easily used in command-line pipelines.

## How to Run

There are binarys available for linux, windows and mac on the github release page.

## Building

To build the tool from source, you will need to have Go installed.

```bash
# Clone the repository (if you haven't already)
# git clone ...

# Navigate to the project directory
cd a2m2a

# Tidy dependencies and build the executable
go mod tidy
go build
```

This will create the `a2m2a` executable in the project directory.

## Usage

The tool can be controlled via several command-line flags.

### Command-Line Flags
-   `-i`, `--in`: Path to the input file. If omitted, reads from `stdin`.
-   `-o`, `--out`: Path to the output file. If omitted, writes to `stdout`.
-   `-w`, `--width`: Sets the canvas width for parsing (default: `80`).
-   `--png`: Forces PNG generation. This is not strictly necessary if your output filename ends with `.png`.
-   `--thumb <width>`: In addition to the main PNG, also generates a thumbnail of the specified pixel width (e.g., `art_thumb.png`).
-   `--16`: Forces the output to be quantized to the 16-color ANSI palette.

### Examples

#### Text Conversion

The tool automatically detects if the input is ANSI or mIRC and converts to the other format.

```bash
# Convert an ANSI file to mIRC format
./a2m2a -i my_art.ans -o my_art.mrc

# Convert a mIRC file to ANSI format
./a2m2a -i my_art.mrc -o my_art.ans
```

#### PNG Generation

PNG generation is triggered automatically if the output filename ends with `.png`. It fully supports the 99-color mIRC palette.

```bash
# Generate a PNG from an ANSI file
./a2m2a -i my_art.ans -o my_art.png

# Generate a high-fidelity PNG from a 99-color mIRC file
./a2m2a -i 99_color_art.mrc -o my_art.png
```

#### Forcing 16-Color Output

You can force the output to the standard 16-color ANSI palette using the `--16` flag. This is particularly useful when converting a 99-color mIRC file into a standard ANSI file.

```bash
# Convert a 99-color mIRC file to a 16-color ANSI file
./a2m2a --in 99_color_art.mrc --out 16_color_art.ans --16
```

#### Thumbnail Generation

Using the `--thumb` flag generates the main PNG file *and* a corresponding thumbnail.

```bash
# Generate a full-size PNG and a 320px wide thumbnail
./a2m2a -i my_art.ans -o my_art.png --thumb 320
# This command creates two files: my_art.png and my_art_thumb.png
```

#### Using with Pipes

The tool fully supports standard I/O. For image generation, you must specify an output file with `-o`.

```bash
# Convert from stdin and print ANSI to stdout
cat my_art.mrc | ./a2m2a

# Generate a PNG from stdin (requires -o)
cat my_art.ans | ./a2m2a -o my_art.png
``` 