package style

// Painted versions of the shared primitives. Card and Window call these when
// the theme's look is watercolor, so every component built on them is
// painted without knowing about it.

import (
	"fmt"
	"math"
	"strings"
)

// Wash densities. Theme-coloured washes stay light so ink text reads on them
// and overlapping components glaze instead of going muddy; a colour the user
// asked for is laid on much heavier, so it still reads as that colour.
const (
	densityAccent     = 0.17
	densityAccentDk   = 0.26
	densityUser       = 0.7
	densityChrome     = 0.15
	densityChip       = 0.3
	densityCode       = 0.9
	inkWidth          = 1.2
	inkWidthFine      = 1.0
	washMisregister   = 2.6
	washInsetPx       = 1.2
	chipInkShade      = 0.35
	controlDotDensity = 0.8
)

// SurfacePaint returns the pigment and density a painted surface uses: the
// user's colour heavily when given, otherwise the accent's pigment lightly.
func SurfacePaint(bg, accent string, t Theme) (pigment string, density float64) {
	if strings.TrimSpace(bg) != "" {
		return Pigment(bg), densityUser
	}
	d := densityAccent
	if t.Name == Dark.Name {
		d = densityAccentDk
	}
	return Pigment(Or(accent, pigmentNeutral)), d
}

// washRect paints a rounded rectangle wash a little off-register from the
// ink line that will be drawn over it.
func washRect(x, y, w, h, r float64, pigment string, density float64, seed uint64, t Theme) string {
	// Each side lands somewhere between a little short of the line and a
	// little past it, never the same on all four.
	rr := newRng(seed ^ 0x5eed)
	m := Clamp(math.Min(w, h)*0.06, 1.5, washMisregister)
	side := func() float64 { return -washInsetPx + rr.signed()*m }
	l, tp, rt, bt := side(), side(), side(), side()
	step := Clamp(math.Max(w, h)/6, 18, 46)
	poly := RoundedRect(x-l, y-tp, w+l+rt, h+tp+bt, r, step)
	maxOff := Clamp(math.Min(w, h)*0.1, 2, 8)
	return Wash(WashSpec{Poly: poly, Pigment: pigment, Density: density, Seed: seed, Layers: 4, Wander: 0.1, MaxOff: maxOff, Settle: 0.6}, t)
}

// inkRect draws a rounded rectangle as one closing pen loop.
func inkRect(x, y, w, h, r, width float64, color string, seed uint64) string {
	poly := RoundedRect(x, y, w, h, r, 12)
	return InkLoop(poly, InkSpec{Width: width, Color: color, Seed: seed})
}

// LineColor is the pen colour for a shape washed with pigment: the same
// pigment mixed much darker (lighter on a dark sheet), so line and wash read
// as one colour family instead of a black outline around a tint.
func LineColor(pigment string, t Theme) string {
	if _, ok := rgbOf(pigment); !ok {
		return t.Ink
	}
	if t.Name == Dark.Name {
		return mix(Shade(pigment, -0.55), t.Ink, 0.35)
	}
	return mix(Shade(pigment, 0.48), t.Ink, 0.18)
}

func mix(a, b string, k float64) string {
	ca, okA := rgbOf(a)
	cb, okB := rgbOf(b)
	if !okA || !okB {
		return a
	}
	var out [3]float64
	for i := range out {
		out[i] = ca[i] + (cb[i]-ca[i])*k
	}
	return hexOf(out)
}

func inkGroup(content string) string {
	if content == "" {
		return ""
	}
	return `<g class="nagare-ink" filter="url(#nagare-wc-ink)">` + content + `</g>`
}

func cardPainted(s CardSpec) string {
	t := s.Theme
	pigment, density := SurfacePaint(s.Bg, s.Accent, t)
	fill := PaintedColor(pigment, density, t)
	titleColor, mutedColor, mutedOpacity := TextColors(s.Fg, fill, t)
	accent := Pigment(Or(s.Accent, t.Muted))
	r := Radius(s.W, s.H)

	s.surface = fill

	var b strings.Builder
	fmt.Fprintf(&b, `<g class="nagare-card %s" transform="translate(%s, %s)">`, s.Class, F(s.X), F(s.Y))
	b.WriteString(washRect(0, 0, s.W, s.H, r, pigment, density, t.Seed, t))
	b.WriteString(inkGroup(inkRect(0.75, 0.75, s.W-1.5, s.H-1.5, r, inkWidth, LineColor(pigment, t), t.Seed+1)))

	pad := Clamp(math.Min(s.W, s.H)*0.1, 4, 14)
	lines := make([]string, 0, 2)
	for _, l := range []string{s.Subtitle, s.Meta} {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, l)
		}
	}
	if s.H < 84 || s.W > s.H*2.4 {
		b.WriteString(cardRow(s, pad, lines, accent, titleColor, mutedColor, mutedOpacity))
	} else {
		b.WriteString(cardStack(s, pad, lines, accent, titleColor, mutedColor, mutedOpacity))
	}
	b.WriteString(`</g>`)
	return b.String()
}

// iconChipPainted lays a small round wash of the accent and draws the icon
// over it in a darker shade of the same pigment, like a field-guide plate.
//
// surface is the colour already painted under the chip. On a dark surface a
// transparent wash would disappear, so the chip is laid as body colour and the
// icon drawn light.
func iconChipPainted(name, accent string, x, y, size float64, surface string, t Theme) string {
	body, ok := icons[name]
	if !ok {
		return ""
	}
	pigment := Pigment(accent)
	seed := t.Seed ^ 0xc41b
	poly := Ellipse(x+size/2, y+size/2, size*0.5, size*0.47, 14)
	spec := WashSpec{Poly: poly, Pigment: pigment, Density: densityChip, Seed: seed, Layers: 5, Wander: 0.14, MaxOff: size * 0.08}
	line := Shade(pigment, chipInkShade)
	if t.Name == Dark.Name {
		line = Shade(pigment, -0.45)
	}
	if l, ok := luminance(surface); ok && l < 0.2 {
		chip := Shade(pigment, -0.2)
		spec.Blend, spec.Density, spec.Pigment = "normal", 0.85, chip
		dark, light := Shade(pigment, 0.65), Shade(pigment, -0.85)
		line = light
		if Contrast(dark, chip) > Contrast(light, chip) {
			line = dark
		}
	}
	var b strings.Builder
	b.WriteString(Wash(spec, t))
	pad := size * 0.2
	scale := (size - 2*pad) / 24
	fmt.Fprintf(&b, `<g transform="translate(%s,%s) scale(%s)" fill="none" stroke="%s" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" filter="url(#nagare-wc-ink)">%s</g>`,
		F(x+pad), F(y+pad), fmt.Sprintf("%.4f", scale), line, body)
	return b.String()
}

// windowPainted paints a window: a wash over the title bar, the content area
// left nearly as bare paper (or washed when it has its own colour), one ink
// loop around the whole and one line under the bar.
func windowPainted(s WindowSpec) string {
	t := s.Theme
	chromePigment, chromeDensity := Pigment(Or(s.Bg, pigmentPaynes)), densityChrome
	if strings.TrimSpace(s.Bg) != "" {
		chromeDensity = densityUser
	}
	chromeFill := PaintedColor(chromePigment, chromeDensity, t)
	titleColor := Legible(s.Fg, chromeFill, t.Muted)
	titleOpacity := 0.0
	if titleColor != t.Muted {
		titleOpacity = 0.9
	}
	r := Radius(s.W, s.H)

	var b strings.Builder
	fmt.Fprintf(&b, `<g class="nagare-window %s" transform="translate(%s, %s)">`, s.Class, F(s.X), F(s.Y))

	// Content area first, so the bar's wash glazes over its top edge.
	contentPigment, contentDensity := contentPaint(s.ContentFill, t)
	if contentDensity > 0 {
		b.WriteString(washRect(s.CX, s.CY, s.CW, s.CH, Clamp(r-3, 2, 7), contentPigment, contentDensity, t.Seed+7, t))
	}
	// The bar: a band of wash that reaches a little past the line under it,
	// as a brush would.
	barPoly := RoundedRect(washInsetPx, washInsetPx, s.W-2*washInsetPx, s.BarH+math.Min(4, s.BarH*0.2), math.Min(r, s.BarH/2), Clamp(s.W/6, 18, 46))
	b.WriteString(Wash(WashSpec{Poly: barPoly, Pigment: chromePigment, Density: chromeDensity, Seed: t.Seed + 3, MaxOff: Clamp(s.BarH*0.12, 1.5, 4)}, t))
	// Below the bar, a much lighter glaze of the same pigment frames the
	// content, so the window still reads as one object.
	if strings.TrimSpace(s.Bg) != "" {
		body := RoundedRect(washInsetPx, s.BarH, s.W-2*washInsetPx, s.H-s.BarH-washInsetPx, math.Min(r, 8), Clamp(s.W/6, 18, 46))
		b.WriteString(Wash(WashSpec{Poly: body, Pigment: chromePigment, Density: chromeDensity * 0.55, Seed: t.Seed + 4, MaxOff: 3}, t))
	}

	var ink strings.Builder
	line := LineColor(chromePigment, t)
	ink.WriteString(inkRect(0.75, 0.75, s.W-1.5, s.H-1.5, r, inkWidth, line, t.Seed+1))
	// The line under the bar does not quite reach the frame on both sides.
	rr := newRng(t.Seed ^ 0xba5)
	lx0 := s.CX + 1 + rr.float()*3
	lx1 := s.CX + s.CW - 1 - rr.float()*3
	ink.WriteString(InkStroke([]Pt{{lx0, s.CY}, {lx1, s.CY + rr.signed()*0.6}}, InkSpec{Width: inkWidthFine * 1.1, Color: line, Seed: t.Seed + 5, TaperIn: 8, TaperOut: 10}))

	// Controls: three dots of pigment.
	cr := Clamp(s.BarH*0.14, 2, 5.5)
	step := cr * 3.1
	ccx := Clamp(s.BarH*0.42, cr+4, 22)
	ccy := s.BarH / 2
	for i, c := range []string{pigmentRose, pigmentYellow, pigmentSap} {
		poly := Ellipse(ccx+float64(i)*step, ccy, cr, cr*0.94, 10)
		b.WriteString(Wash(WashSpec{Poly: poly, Pigment: c, Density: controlDotDensity, Seed: t.Seed + 20 + uint64(i), Layers: 4, Wander: 0.12, MaxOff: cr * 0.18, Blend: "normal", Filter: "none"}, t))
	}
	controlsEnd := ccx + 2*step + cr

	size := Clamp(s.BarH*0.36, 8, 13)
	left := controlsEnd + size
	right := s.W - left
	var text string
	if right-left > size*3 && s.Title != "" {
		if s.TitlePill {
			ph := Clamp(s.BarH*0.56, size+4, 26)
			pill := RoundedRect(left, ccy-ph/2, right-left, ph, ph/2, 10)
			ink.WriteString(InkLoop(pill, InkSpec{Width: inkWidthFine, Color: line, Seed: t.Seed + 6, Opacity: 0.75}))
			text = Text(s.Title, TextOpts{X: left + ph*0.55, Y: ccy, Size: size, Color: titleColor, Opacity: titleOpacity, Mono: s.TitleMono, Width: right - left - ph*1.1})
		} else {
			text = Text(s.Title, TextOpts{X: s.W / 2, Y: ccy, Size: size, Color: titleColor, Opacity: titleOpacity, Anchor: "middle", Weight: "600", Mono: s.TitleMono, Width: right - left})
		}
	}
	b.WriteString(inkGroup(ink.String()))
	b.WriteString(text)
	return b.String()
}

// contentPaint decides how a window's content area is painted. The theme's
// own surfaces are left as paper (a whisper of umber for an inset); any other
// colour is a real wash.
func contentPaint(fill string, t Theme) (string, float64) {
	switch strings.TrimSpace(fill) {
	case "", t.Surface, t.Paper:
		return "", 0
	case t.Inset:
		return "", 0
	case t.CodeBg:
		return t.CodeBg, densityCode
	}
	return Pigment(fill), densityUser
}
