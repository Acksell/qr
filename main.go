// Command qr encodes text into QR code images and decodes QR codes back out of
// images from the command line.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"io"
	"os"
	"strings"

	"golang.design/x/clipboard"

	"github.com/acksell/qr/internal/qr"
)

const usage = `qr — encode and decode QR codes

Usage:
  qr                            Read the clipboard: decode an image's QR code,
                                 or encode clipboard text into one
  qr encode [flags] [text]     Generate a QR code from text (reads stdin if text is omitted)
  qr decode [flags] <image>    Detect and decode a QR code from an image file

Run "qr encode -h" or "qr decode -h" for command-specific flags.`

func main() {
	var err error
	switch {
	case len(os.Args) < 2:
		err = runClipboard()
	case os.Args[1] == "encode" || os.Args[1] == "enc" || os.Args[1] == "e":
		err = runEncode(os.Args[2:])
	case os.Args[1] == "decode" || os.Args[1] == "dec" || os.Args[1] == "d":
		err = runDecode(os.Args[2:])
	case os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "help":
		fmt.Println(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s\n", os.Args[1], usage)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// runClipboard inspects the system clipboard: an image is scanned for a QR
// code and its text printed, while text is rendered into a QR code and
// printed to the terminal alongside the text itself.
func runClipboard() error {
	if err := clipboard.Init(); err != nil {
		return fmt.Errorf("clipboard unavailable: %w", err)
	}

	if buf := clipboard.Read(clipboard.FmtImage); len(buf) > 0 {
		img, _, err := image.Decode(bytes.NewReader(buf))
		if err != nil {
			return fmt.Errorf("decoding clipboard image: %w", err)
		}
		res, err := qr.DecodeImage(img)
		if err != nil {
			return err
		}
		printDecoded(res)
		return nil
	}

	if buf := clipboard.Read(clipboard.FmtText); len(strings.TrimSpace(string(buf))) > 0 {
		text := strings.TrimSpace(string(buf))
		lvl, err := qr.ParseLevel("")
		if err != nil {
			return err
		}
		s, err := qr.EncodeString(text, lvl, false)
		if err != nil {
			return err
		}
		fmt.Println(text)
		fmt.Println()
		fmt.Print(s)
		return nil
	}

	return fmt.Errorf("clipboard has no image or text to work with")
}

func runEncode(args []string) error {
	fs := flag.NewFlagSet("encode", flag.ContinueOnError)
	output := fs.String("output", "qr.png", "output PNG file path (\"-\" writes PNG to stdout)")
	size := fs.Int("size", 512, "output image width/height in pixels")
	level := fs.String("level", "medium", "error-correction level: low|medium|high|highest")
	terminal := fs.Bool("terminal", false, "print the QR code to the terminal instead of writing a file")
	invert := fs.Bool("invert", false, "with --terminal, swap colors (helps scanning on dark backgrounds)")

	// Short aliases bound to the same variables.
	fs.StringVar(output, "o", *output, "shorthand for --output")
	fs.IntVar(size, "s", *size, "shorthand for --size")
	fs.StringVar(level, "l", *level, "shorthand for --level")
	fs.BoolVar(terminal, "t", *terminal, "shorthand for --terminal")

	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Generate a QR code from text.\n\nUsage:\n  qr encode [flags] [text]\n\nFlags:")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	text, err := encodeInput(fs.Args())
	if err != nil {
		return err
	}

	lvl, err := qr.ParseLevel(*level)
	if err != nil {
		return err
	}

	if *terminal {
		s, err := qr.EncodeString(text, lvl, *invert)
		if err != nil {
			return err
		}
		fmt.Print(s)
		return nil
	}

	png, err := qr.EncodePNG(text, lvl, *size)
	if err != nil {
		return err
	}

	if *output == "-" {
		_, err := os.Stdout.Write(png)
		return err
	}
	if err := os.WriteFile(*output, png, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", *output, err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d bytes)\n", *output, len(png))
	return nil
}

// encodeInput resolves the text to encode: joined positional args, or stdin
// when no args are given.
func encodeInput(args []string) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("reading stdin: %w", err)
	}
	text := strings.TrimRight(string(data), "\r\n")
	if text == "" {
		return "", fmt.Errorf("no text to encode: pass it as an argument or pipe it via stdin")
	}
	return text, nil
}

func runDecode(args []string) error {
	fs := flag.NewFlagSet("decode", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "output the decoded text and metadata as JSON")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Detect and decode a QR code from an image (PNG, JPEG or GIF).\n\nUsage:\n  qr decode [flags] <image>   (use \"-\" to read the image from stdin)\n\nFlags:")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("decode needs exactly one image path (got %d)", fs.NArg())
	}

	img, err := openImage(fs.Arg(0))
	if err != nil {
		return err
	}

	res, err := qr.DecodeImage(img)
	if err != nil {
		return err
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(res)
	}
	printDecoded(res)
	return nil
}

// printDecoded reports a successful decode, making it unambiguous (versus a
// blank or unusual result) that a QR code was actually found.
func printDecoded(res *qr.Result) {
	fmt.Printf("\033[32m✓\033[0m qr code says: %s\n", res.Text)
}

// openImage decodes an image from a file path, or from stdin when path is "-".
// The PNG/JPEG/GIF format decoders are registered by the qr package.
func openImage(path string) (image.Image, error) {
	in := os.Stdin
	if path != "-" {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		in = f
	}
	img, _, err := image.Decode(in)
	if err != nil {
		return nil, fmt.Errorf("decoding image: %w (supported: png, jpeg, gif)", err)
	}
	return img, nil
}
