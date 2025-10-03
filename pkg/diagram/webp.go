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

type affineTransform struct {
	a, b, c, d, e, f float64
}

func identityTransform() affineTransform {
	return affineTransform{a: 1, d: 1}
}

func (t affineTransform) Multiply(o affineTransform) affineTransform {
	return affineTransform{
		a: t.a*o.a + t.c*o.b,
		b: t.b*o.a + t.d*o.b,
		c: t.a*o.c + t.c*o.d,
		d: t.b*o.c + t.d*o.d,
		e: t.a*o.e + t.c*o.f + t.e,
		f: t.b*o.e + t.d*o.f + t.f,
	}
}

func (t affineTransform) Apply(x, y float64) (float64, float64) {
	return t.a*x + t.c*y + t.e, t.b*x + t.d*y + t.f
}

func (t affineTransform) ScaleFactor() float64 {
	sx := math.Hypot(t.a, t.b)
	sy := math.Hypot(t.c, t.d)
	if sx == 0 && sy == 0 {
		return 1
	}
	if sx == 0 {
		return sy
	}
	if sy == 0 {
		return sx
	}
	return (sx + sy) / 2
}

func extractTextElements(svg string) ([]textElement, error) {
	decoder := xml.NewDecoder(strings.NewReader(svg))
	var elements []textElement

	transformStack := []affineTransform{identityTransform()}
	var current *textElement
	var content strings.Builder
	textDepth := 0

	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			parent := transformStack[len(transformStack)-1]
			combined := parent
			if tr := getAttr(t.Attr, "transform"); tr != "" {
				combined = parent.Multiply(parseTransformAttribute(tr))
			}
			transformStack = append(transformStack, combined)

			if current != nil {
				textDepth++
				continue
			}

			if t.Name.Local != "text" {
				continue
			}

			element := textElement{
				FontSize: 16,
				Fill:     color.Black,
			}
			var x, y float64
			for _, attr := range t.Attr {
				switch attr.Name.Local {
				case "x":
					x = parseSVGFloat(splitFirstValue(attr.Value), 0)
				case "y":
					y = parseSVGFloat(splitFirstValue(attr.Value), 0)
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

			element.FontSize *= combined.ScaleFactor()
			element.X, element.Y = combined.Apply(x, y)

			current = &element
			content.Reset()
			textDepth = 0
		case xml.CharData:
			if current != nil {
				content.WriteString(string(t))
			}
		case xml.EndElement:
			if current != nil {
				if textDepth > 0 {
					textDepth--
				} else if t.Name.Local == "text" {
					current.Text = strings.TrimSpace(html.UnescapeString(content.String()))
					if current.Text != "" {
						elements = append(elements, *current)
					}
					current = nil
				}
			}
			if len(transformStack) > 1 {
				transformStack = transformStack[:len(transformStack)-1]
			}
		}
	}

	return elements, nil
}

func getAttr(attrs []xml.Attr, name string) string {
	for _, attr := range attrs {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}

func splitFirstValue(v string) string {
	fields := strings.FieldsFunc(v, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\n' || r == '\t' || r == '\r'
	})
	if len(fields) == 0 {
		return v
	}
	return fields[0]
}

func parseTransformAttribute(value string) affineTransform {
	result := identityTransform()
	s := strings.TrimSpace(value)
	for len(s) > 0 {
		open := strings.IndexByte(s, '(')
		if open == -1 {
			break
		}
		name := strings.TrimSpace(s[:open])
		s = s[open+1:]
		close := strings.IndexByte(s, ')')
		if close == -1 {
			break
		}
		args := s[:close]
		s = strings.TrimSpace(s[close+1:])

		params := parseNumberList(args)
		switch name {
		case "translate":
			dx, dy := 0.0, 0.0
			if len(params) > 0 {
				dx = params[0]
			}
			if len(params) > 1 {
				dy = params[1]
			}
			result = result.Multiply(affineTransform{a: 1, d: 1, e: dx, f: dy})
		case "scale":
			sx, sy := 1.0, 1.0
			if len(params) > 0 {
				sx = params[0]
			}
			if len(params) > 1 {
				sy = params[1]
			} else {
				sy = sx
			}
			result = result.Multiply(affineTransform{a: sx, d: sy})
		case "matrix":
			if len(params) >= 6 {
				result = result.Multiply(affineTransform{
					a: params[0],
					b: params[1],
					c: params[2],
					d: params[3],
					e: params[4],
					f: params[5],
				})
			}
		}
	}
	return result
}

func parseNumberList(value string) []float64 {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\n' || r == '\t' || r == '\r'
	})
	nums := make([]float64, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		nums = append(nums, parseSVGFloat(part, 0))
	}
	return nums
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
