package diagram

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html"
	"image"
	"image/color"
	"image/draw"
	"io"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/chai2010/webp"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
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

	texts, err := extractTextElements(svg)
	if err != nil {
		return nil, fmt.Errorf("extract text: %w", err)
	}
	if err := drawTextElements(rgba, texts); err != nil {
		return nil, fmt.Errorf("draw text: %w", err)
	}

	buf := bytes.NewBuffer(nil)
	if err := webp.Encode(buf, rgba, &webp.Options{Lossless: true}); err != nil {
		return nil, fmt.Errorf("encode webp: %w", err)
	}

	return buf.Bytes(), nil
}

type textElement struct {
	X                float64
	Y                float64
	Anchor           string
	DominantBaseline string
	FontSize         float64
	Fill             color.Color
	Text             string
}

func extractTextElements(svg string) ([]textElement, error) {
	decoder := xml.NewDecoder(strings.NewReader(svg))
	var elements []textElement

	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		start, ok := tok.(xml.StartElement)
		if !ok || start.Name.Local != "text" {
			continue
		}

		element := textElement{
			FontSize: 16,
			Fill:     color.Black,
		}

		for _, attr := range start.Attr {
			switch attr.Name.Local {
			case "x":
				element.X = parseSVGFloat(attr.Value, 0)
			case "y":
				element.Y = parseSVGFloat(attr.Value, 0)
			case "font-size":
				element.FontSize = parseSVGFloat(attr.Value, element.FontSize)
			case "text-anchor":
				element.Anchor = attr.Value
			case "dominant-baseline":
				element.DominantBaseline = attr.Value
			case "fill":
				if col, err := parseSVGColor(attr.Value); err == nil {
					element.Fill = col
				}
			}
		}

		var content strings.Builder
		depth := 1
		for depth > 0 {
			tok, err := decoder.Token()
			if err != nil {
				return nil, err
			}

			switch v := tok.(type) {
			case xml.CharData:
				content.WriteString(string(v))
			case xml.StartElement:
				depth++
			case xml.EndElement:
				depth--
			}
		}

		element.Text = strings.TrimSpace(html.UnescapeString(content.String()))
		if element.Text == "" {
			continue
		}

		elements = append(elements, element)
	}

	return elements, nil
}

var (
	fontOnce     sync.Once
	fontInitErr  error
	regularFont  *opentype.Font
	fontFaceLock sync.Mutex
	fontFaces    = make(map[float64]font.Face)
)

func ensureFontLoaded() error {
	fontOnce.Do(func() {
		regularFont, fontInitErr = opentype.Parse(goregular.TTF)
	})
	return fontInitErr
}

func getFontFace(size float64) (font.Face, error) {
	if err := ensureFontLoaded(); err != nil {
		return nil, err
	}

	fontFaceLock.Lock()
	defer fontFaceLock.Unlock()

	if face, ok := fontFaces[size]; ok {
		return face, nil
	}

	face, err := opentype.NewFace(regularFont, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		return nil, err
	}

	fontFaces[size] = face
	return face, nil
}

func drawTextElements(dst *image.RGBA, elements []textElement) error {
	for _, element := range elements {
		face, err := getFontFace(element.FontSize)
		if err != nil {
			return err
		}

		d := font.Drawer{
			Dst:  dst,
			Src:  image.NewUniform(element.Fill),
			Face: face,
		}

		width := d.MeasureString(element.Text)
		x := element.X
		switch element.Anchor {
		case "middle":
			x -= float64(width) / (64 * 2)
		case "end":
			x -= float64(width) / 64
		}

		y := element.Y
		if strings.EqualFold(element.DominantBaseline, "middle") {
			metrics := face.Metrics()
			y -= float64(metrics.Descent-metrics.Ascent) / (64 * 2)
		}

		d.Dot = fixed.Point26_6{
			X: fixed.Int26_6(math.Round(x * 64)),
			Y: fixed.Int26_6(math.Round(y * 64)),
		}
		d.DrawString(element.Text)
	}

	return nil
}

func parseSVGFloat(value string, fallback float64) float64 {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, "px")
	if value == "" {
		return fallback
	}
	if v, err := strconv.ParseFloat(value, 64); err == nil {
		return v
	}
	return fallback
}

func parseSVGColor(value string) (color.Color, error) {
	v := strings.TrimSpace(strings.ToLower(value))
	switch {
	case strings.HasPrefix(v, "#"):
		return parseHexColor(v)
	case strings.HasPrefix(v, "rgb"):
		return parseRGBColor(v)
	default:
		return parseNamedColor(v)
	}
}

func parseHexColor(v string) (color.Color, error) {
	v = strings.TrimPrefix(v, "#")
	if len(v) == 3 {
		return color.RGBA{
			R: duplicateHex(v[0]),
			G: duplicateHex(v[1]),
			B: duplicateHex(v[2]),
			A: 255,
		}, nil
	}
	if len(v) == 6 {
		r, err := strconv.ParseUint(v[0:2], 16, 8)
		if err != nil {
			return nil, err
		}
		g, err := strconv.ParseUint(v[2:4], 16, 8)
		if err != nil {
			return nil, err
		}
		b, err := strconv.ParseUint(v[4:6], 16, 8)
		if err != nil {
			return nil, err
		}
		return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
	}
	return nil, fmt.Errorf("unsupported hex color %q", v)
}

func duplicateHex(b byte) uint8 {
	return uint8((hexToInt(b) << 4) | hexToInt(b))
}

func hexToInt(b byte) uint8 {
	switch {
	case b >= '0' && b <= '9':
		return uint8(b - '0')
	case b >= 'a' && b <= 'f':
		return uint8(b-'a') + 10
	case b >= 'A' && b <= 'F':
		return uint8(b-'A') + 10
	default:
		return 0
	}
}

func parseRGBColor(v string) (color.Color, error) {
	start := strings.IndexRune(v, '(')
	end := strings.IndexRune(v, ')')
	if start < 0 || end < 0 || end <= start+1 {
		return nil, fmt.Errorf("invalid rgb color %q", v)
	}
	parts := strings.Split(v[start+1:end], ",")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid rgb color %q", v)
	}

	vals := make([]uint8, 3)
	for i := 0; i < 3; i++ {
		p := strings.TrimSpace(parts[i])
		n, err := strconv.Atoi(strings.TrimSuffix(p, "%"))
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(p, "%") {
			n = int(math.Round(float64(n) * 255.0 / 100.0))
		}
		if n < 0 {
			n = 0
		}
		if n > 255 {
			n = 255
		}
		vals[i] = uint8(n)
	}

	return color.RGBA{R: vals[0], G: vals[1], B: vals[2], A: 255}, nil
}

func parseNamedColor(v string) (color.Color, error) {
	switch v {
	case "black":
		return color.Black, nil
	case "white":
		return color.White, nil
	case "red":
		return color.RGBA{R: 255, A: 255}, nil
	case "green":
		return color.RGBA{G: 128, A: 255}, nil
	case "blue":
		return color.RGBA{B: 255, A: 255}, nil
	case "transparent":
		return color.RGBA{}, nil
	default:
		return color.Black, fmt.Errorf("unsupported color name %q", v)
	}
}
