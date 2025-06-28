# a2m2a - ANSI / mIRC Art Converter

`a2m2a` (ANSI to mIRC Art to ANSI) is a versatile command-line tool written in Go for converting legacy text-based art formats and rendering them as images. It can seamlessly convert between 16-color ANSI art and mIRC art, and can also generate high-quality PNG images from either format.

Inspired by an original C utility `a2m` by t4t3r, `a2m2a` has been modernized with features like automatic format detection and image rendering capabilities.

## Features

-   **Two-Way Text Conversion:**
    -   Convert ANSI art (`.ans`) to mIRC art (`.mrc`).
    -   Convert mIRC art (`.mrc`) to ANSI art (`.ans`).
-   **Automatic Format Detection:** No need to specify the input format; the tool inspects the file and determines the correct conversion path automatically.
-   **PNG Image Generation:**
    -   Render either ANSI or mIRC art directly to a PNG file.
    -   Creates tightly-cropped images based on the artwork's content.
-   **Thumbnail Generation:**
    -   Create a smaller thumbnail of the artwork with a user-specified width.
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
-   `--png`: Renders the output as a PNG image.
-   `--thumb <width>`: Renders the output as a PNG thumbnail of the specified pixel width.
-   `-w <width>`: Sets the canvas width for parsing (default: `80`).

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

```bash
# Generate a PNG from an ANSI file
./a2m2a -i my_art.ans --png -o my_art.png

# Generate a PNG from a mIRC file
./a2m2a -i my_art.mrc --png -o my_art.png
```

#### Thumbnail Generation

```bash
# Generate a 320px wide thumbnail from an ANSI file
./a2m2a -i my_art.ans --thumb 320 -o my_art_thumb.png
```

#### Using with Pipes

The tool fully supports standard I/O.

```bash
# Convert from stdin and print ANSI to stdout
cat my_art.mrc | ./a2m2a

# Generate a PNG from stdin
cat my_art.ans | ./a2m2a --png > my_art.png
``` 