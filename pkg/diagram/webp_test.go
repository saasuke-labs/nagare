package diagram

import (
	"bytes"
	"image/color"
	"testing"

	"github.com/chai2010/webp"
)

func TestRasterizeSVGToWebPIncludesText(t *testing.T) {
	const svg = `<?xml version="1.0" encoding="UTF-8"?>
<svg width="120" height="40" viewBox="0 0 120 40" xmlns="http://www.w3.org/2000/svg">
        <rect x="0" y="0" width="120" height="40" fill="#ffffff" />
        <text x="60" y="20" text-anchor="middle" dominant-baseline="middle" font-size="18" fill="#ff0000">Hi</text>
</svg>`

	data, err := rasterizeSVGToWebP(svg, 120, 40)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	img, err := webp.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode webp: %v", err)
	}

	bounds := img.Bounds()
	hasTextPixel := false
	for y := bounds.Min.Y; y < bounds.Max.Y && !hasTextPixel; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if c.A != 0 && !(c.R == 255 && c.G == 255 && c.B == 255) {
				hasTextPixel = true
				break
			}
		}
	}

	if !hasTextPixel {
		t.Fatalf("expected non-white pixels for text rendering")
	}
}
