package qr

import (
	"bytes"
	"image/png"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	cases := []string{
		"hello world",
		"https://example.com/path?q=1&x=2",
		"unicode: åäö 日本語 🚀",
		"1234567890",
	}

	for _, want := range cases {
		t.Run(want, func(t *testing.T) {
			pngBytes, err := EncodePNG(want, mustLevel(t, "medium"), 256)
			if err != nil {
				t.Fatalf("EncodePNG: %v", err)
			}

			img, err := png.Decode(bytes.NewReader(pngBytes))
			if err != nil {
				t.Fatalf("png.Decode: %v", err)
			}

			res, err := DecodeImage(img)
			if err != nil {
				t.Fatalf("DecodeImage: %v", err)
			}
			if res.Text != want {
				t.Errorf("round trip mismatch:\n got %q\nwant %q", res.Text, want)
			}
			if res.Format != "QR_CODE" {
				t.Errorf("format = %q, want QR_CODE", res.Format)
			}
		})
	}
}

func TestEncodeEmpty(t *testing.T) {
	if _, err := EncodePNG("", mustLevel(t, "low"), 256); err == nil {
		t.Fatal("expected error encoding empty text, got nil")
	}
}

func TestParseLevel(t *testing.T) {
	valid := []string{"", "low", "L", "medium", "m", "high", "Q", "highest", "h"}
	for _, s := range valid {
		if _, err := ParseLevel(s); err != nil {
			t.Errorf("ParseLevel(%q) unexpected error: %v", s, err)
		}
	}
	if _, err := ParseLevel("nonsense"); err == nil {
		t.Error("ParseLevel(\"nonsense\") = nil error, want error")
	}
}

func mustLevel(t *testing.T, s string) Level {
	t.Helper()
	l, err := ParseLevel(s)
	if err != nil {
		t.Fatalf("ParseLevel(%q): %v", s, err)
	}
	return l
}
