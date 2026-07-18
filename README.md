# qr

A small command-line tool to **encode** text into QR code images and **decode**
QR codes back out of images.

- Encoding uses [`skip2/go-qrcode`](https://github.com/skip2/go-qrcode).
- Decoding uses [`makiuchi-d/gozxing`](https://github.com/makiuchi-d/gozxing)
  (a Go port of ZXing), which reads real-world photos — rotated, skewed, or
  low-contrast codes included.

## Install / build

```sh
go build -o qr .
```

This produces a `qr` binary in the current directory. (`go install .` puts it on
your `$PATH`.)

## Usage

```
qr                            Read the clipboard: decode an image's QR code,
                               or encode clipboard text into one
qr encode [flags] [text]     Generate a QR code from text
qr decode [flags] <image>    Detect and decode a QR code from an image
```

### Clipboard

```sh
qr
```

Run with no arguments and `qr` reads the system clipboard:

- If it holds an **image**, any QR code in it is decoded and the text is
  printed.
- If it holds **text**, that text is printed back along with a QR code
  rendered for the terminal.

### Encode

```sh
# Write a PNG (default: qr.png)
qr encode -o site.png "https://example.com"

# Read the text from stdin
echo "hello from stdin" | qr encode -o hello.png

# Print PNG bytes to stdout (pipe it somewhere)
qr encode -o - "data" > code.png

# Render straight to the terminal (no file written)
qr encode --terminal "scan me"
qr encode --terminal --invert "scan me"   # for dark-background terminals
```

Encode flags:

| Flag | Alias | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `qr.png` | Output PNG path (`-` writes to stdout) |
| `--size` | `-s` | `512` | Image width/height in pixels |
| `--level` | `-l` | `medium` | Error correction: `low`, `medium`, `high`, `highest` |
| `--terminal` | `-t` | `false` | Print to the terminal instead of a file |
| `--invert` | | `false` | With `--terminal`, swap colors |

### Decode

```sh
# Print the decoded text
qr decode site.png

# Structured output with format + finder-point coordinates
qr decode --json site.png

# Read the image from stdin
cat site.png | qr decode -
```

Supported input image formats: **PNG, JPEG, GIF**.

## Example round trip

```sh
qr encode -o /tmp/x.png "round trip"
qr decode /tmp/x.png
# -> round trip
```

## Exit codes

- `0` — success
- `1` — runtime error (no QR found, unreadable image, invalid flag value, …)
- `2` — usage error (unknown command, missing arguments)

## Test

```sh
go test ./...
```
