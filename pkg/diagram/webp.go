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
	"unicode"
	"unicode/utf8"

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

	if err := drawSVGText(rgba, svg); err != nil {
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
	Fill             color.Color
	Opacity          float64
	FontSize         float64
	TextAnchor       string
	DominantBaseline string
	Content          string
}

func drawSVGText(img *image.RGBA, svg string) error {
	elements, err := extractTextElements(svg)
	if err != nil {
		return err
	}

	for _, element := range elements {
		if strings.TrimSpace(element.Content) == "" {
			continue
		}
		if element.Opacity <= 0 {
			continue
		}

		face, err := getFontFace(element.FontSize)
		if err != nil {
			return fmt.Errorf("load font: %w", err)
		}

		drawer := font.Drawer{Face: face}
		advance := drawer.MeasureString(element.Content)

		x := element.X
		switch element.TextAnchor {
		case "middle":
			x -= float64(advance>>6) / 2
		case "end":
			x -= float64(advance >> 6)
		}

		metrics := face.Metrics()
		baseline := element.Y
		ascent := float64(metrics.Ascent) / 64
		descent := float64(metrics.Descent) / 64
		switch element.DominantBaseline {
		case "middle", "central":
			baseline += (ascent - descent) / 2
		case "hanging":
			baseline += ascent * 0.8
		case "text-after-edge", "ideographic":
			baseline -= descent
		}

		drawer = font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(applyOpacity(element.Fill, element.Opacity)),
			Face: face,
			Dot:  fixedPoint(x, baseline),
		}
		drawer.DrawString(element.Content)
	}

	return nil
}

func fixedPoint(x, y float64) fixed.Point26_6 {
	return fixed.Point26_6{
		X: fixed.Int26_6(math.Round(x * 64)),
		Y: fixed.Int26_6(math.Round(y * 64)),
	}
}

func applyOpacity(clr color.Color, opacity float64) color.Color {
	if opacity >= 1.0 {
		return clr
	}
	if opacity <= 0 {
		return color.NRGBA{0, 0, 0, 0}
	}
	r, g, b, a := clr.RGBA()
	alpha := float64(a>>8) * opacity
	if alpha > 255 {
		alpha = 255
	}
	return color.NRGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(alpha + 0.5),
	}
}

var (
	fontOnce sync.Once
	fontFace *opentype.Font
	fontErr  error

	faceCache   = make(map[float64]font.Face)
	faceCacheMu sync.Mutex
)

func getFontFace(size float64) (font.Face, error) {
	if size <= 0 {
		size = 16
	}
	fontOnce.Do(func() {
		fontFace, fontErr = opentype.Parse(goregular.TTF)
	})
	if fontErr != nil {
		return nil, fontErr
	}

	faceCacheMu.Lock()
	defer faceCacheMu.Unlock()

	if face, ok := faceCache[size]; ok {
		return face, nil
	}

	face, err := opentype.NewFace(fontFace, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}
	faceCache[size] = face
	return face, nil
}

func extractTextElements(svg string) ([]textElement, error) {
	decoder := xml.NewDecoder(strings.NewReader(svg))
	var elements []textElement

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}

		if start.Name.Local != "text" {
			continue
		}

		element := textElement{
			Fill:     color.Black,
			Opacity:  1,
			FontSize: 16,
		}
		applyAttributes(&element, start.Attr)

		var contentBuilder strings.Builder
		depth := 1
		for depth > 0 {
			token, err = decoder.Token()
			if err != nil {
				return nil, err
			}

			switch t := token.(type) {
			case xml.StartElement:
				depth++
				if t.Name.Local == "tspan" {
					applyAttributes(&element, t.Attr)
				}
			case xml.EndElement:
				depth--
			case xml.CharData:
				contentBuilder.WriteString(string(t))
			}
		}

		element.Content = normalizeTextContent(contentBuilder.String())
		elements = append(elements, element)
	}

	return elements, nil
}

func applyAttributes(element *textElement, attrs []xml.Attr) {
	for _, attr := range attrs {
		value := strings.TrimSpace(attr.Value)
		switch attr.Name.Local {
		case "x":
			if f, ok := parseNumber(value); ok {
				element.X = f
			}
		case "y":
			if f, ok := parseNumber(value); ok {
				element.Y = f
			}
		case "font-size":
			if f, ok := parseNumber(value); ok {
				element.FontSize = f
			}
		case "text-anchor":
			element.TextAnchor = value
		case "dominant-baseline":
			element.DominantBaseline = value
		case "fill":
			if clr, err := oksvg.ParseSVGColor(value); err == nil {
				if clr == nil {
					element.Opacity = 0
				} else {
					element.Fill = clr
				}
			}
		case "fill-opacity", "opacity":
			if f, ok := parseNumber(value); ok {
				element.Opacity = clamp01(f)
			}
		case "style":
			applyStyleAttribute(element, value)
		}
	}
}

func applyStyleAttribute(element *textElement, style string) {
	declarations := strings.Split(style, ";")
	for _, declaration := range declarations {
		declaration = strings.TrimSpace(declaration)
		if declaration == "" {
			continue
		}
		parts := strings.SplitN(declaration, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "font-size":
			if f, ok := parseNumber(value); ok {
				element.FontSize = f
			}
		case "fill":
			if clr, err := oksvg.ParseSVGColor(value); err == nil {
				if clr == nil {
					element.Opacity = 0
				} else {
					element.Fill = clr
				}
			}
		case "fill-opacity", "opacity":
			if f, ok := parseNumber(value); ok {
				element.Opacity = clamp01(f)
			}
		case "text-anchor":
			element.TextAnchor = value
		case "dominant-baseline":
			element.DominantBaseline = value
		}
	}
}

func parseNumber(value string) (float64, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	value = strings.TrimSuffix(value, "px")
	// take first component if multiple values provided
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return r == ' ' || r == ','
	})
	if len(fields) == 0 {
		return 0, false
	}
	f, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func normalizeTextContent(s string) string {
	if s == "" {
		return ""
	}
	s = html.UnescapeString(s)
	var builder strings.Builder
	prevSpace := false
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		if unicode.IsSpace(r) {
			if prevSpace {
				continue
			}
			r = ' '
			prevSpace = true
		} else {
			prevSpace = false
		}
		builder.WriteRune(r)
	}
	return strings.TrimSpace(builder.String())
}
