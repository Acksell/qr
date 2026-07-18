package qr

import (
	"fmt"
	"image"

	// Register the common image decoders so DecodeImage accepts PNG, JPEG and
	// GIF sources without the caller wiring them up.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

// Point is a corner/finder location reported by the decoder, in image pixels.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Result holds a decoded QR code and where it was found in the image.
type Result struct {
	Text   string  `json:"text"`
	Format string  `json:"format"`
	Points []Point `json:"points"`
}

// DecodeImage locates and decodes a single QR code in img.
func DecodeImage(img image.Image) (*Result, error) {
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, fmt.Errorf("preparing image: %w", err)
	}

	// TRY_HARDER trades speed for a better chance at rotated, skewed or
	// low-contrast codes — worth it for a one-shot CLI decode.
	hints := map[gozxing.DecodeHintType]any{
		gozxing.DecodeHintType_TRY_HARDER: true,
	}

	res, err := qrcode.NewQRCodeReader().Decode(bmp, hints)
	if err != nil {
		return nil, fmt.Errorf("no QR code found in image: %w", err)
	}

	pts := make([]Point, 0, len(res.GetResultPoints()))
	for _, p := range res.GetResultPoints() {
		if p == nil {
			continue
		}
		pts = append(pts, Point{X: p.GetX(), Y: p.GetY()})
	}

	return &Result{
		Text:   res.GetText(),
		Format: res.GetBarcodeFormat().String(),
		Points: pts,
	}, nil
}
