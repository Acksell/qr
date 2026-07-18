package qr

import (
	"fmt"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

// Level is a QR error-correction level. Higher levels tolerate more damage to
// the code at the cost of storing less data in the same footprint.
type Level = qrcode.RecoveryLevel

// ParseLevel maps a human-friendly name (or the standard L/M/Q/H letters) to a
// recovery level. An empty string defaults to Medium.
func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "medium", "m":
		return qrcode.Medium, nil // ~15% recovery
	case "low", "l":
		return qrcode.Low, nil // ~7% recovery
	case "high", "q":
		return qrcode.High, nil // ~25% recovery
	case "highest", "h":
		return qrcode.Highest, nil // ~30% recovery
	default:
		return 0, fmt.Errorf("invalid error-correction level %q (use low|medium|high|highest)", s)
	}
}

// EncodePNG renders text as a QR code and returns the PNG bytes. size is the
// width/height of the output image in pixels.
func EncodePNG(text string, level Level, size int) ([]byte, error) {
	q, err := newCode(text, level)
	if err != nil {
		return nil, err
	}
	return q.PNG(size)
}

// EncodeString renders text as a QR code drawn with Unicode half-block
// characters, suitable for printing to a terminal. When invert is true the
// foreground and background are swapped, which helps the code scan on dark
// terminal backgrounds.
func EncodeString(text string, level Level, invert bool) (string, error) {
	q, err := newCode(text, level)
	if err != nil {
		return "", err
	}
	return q.ToSmallString(invert), nil
}

func newCode(text string, level Level) (*qrcode.QRCode, error) {
	if text == "" {
		return nil, fmt.Errorf("nothing to encode: input text is empty")
	}
	q, err := qrcode.New(text, level)
	if err != nil {
		return nil, fmt.Errorf("encoding QR code: %w", err)
	}
	return q, nil
}
