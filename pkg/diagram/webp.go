package diagram

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"

	"github.com/chai2010/webp"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// CreateDiagramWebP generates a diagram identical to CreateDiagram but returns it encoded as a WebP image.
func CreateDiagramWebP(code string) ([]byte, error) {
	svg, width, height, err := CreateDiagramWithSize(code)
	if err != nil {
		return nil, err
	}

	data, err := rasterizeSVGToWebP(svg, width, height)
	if err != nil {
		return nil, fmt.Errorf("convert to webp: %w", err)
	}

	return data, nil
}

func rasterizeSVGToWebP(svg string, width, height int) ([]byte, error) {
	icon, err := oksvg.ReadIconStream(strings.NewReader(svg))
	if err != nil {
		return nil, fmt.Errorf("parse svg: %w", err)
	}

	// Determine the output dimensions, falling back to the viewBox if necessary.
	box := icon.ViewBox
	if width <= 0 {
		width = int(math.Ceil(box.W))
	}
	if height <= 0 {
		height = int(math.Ceil(box.H))
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid canvas size: %dx%d", width, height)
	}

	icon.SetTarget(0, 0, float64(width), float64(height))

	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(rgba, rgba.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	scanner := rasterx.NewScannerGV(width, height, rgba, rgba.Bounds())
	raster := rasterx.NewDasher(width, height, scanner)
	icon.Draw(raster, 1.0)

	buf := bytes.NewBuffer(nil)
	if err := webp.Encode(buf, rgba, &webp.Options{Lossless: true}); err != nil {
		return nil, fmt.Errorf("encode webp: %w", err)
	}

	return buf.Bytes(), nil
}
