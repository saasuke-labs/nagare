package style

// The watercolor look: components are painted on paper instead of filled.
//
// It follows the painted register worked out for the Zen game and the rocket
// video: a warm sheet, transparent pigment washes built from a few deformed
// layers that glaze where they overlap and pool darker at their drying edge,
// granulation from the paper's tooth, and one pressure-varying ink line drawn
// over the wash a little off-register, the way a hand would.
//
// Everything is baked into geometry from a per-node seed, so a diagram renders
// the same way every time. Three small shared SVG filters (paper, wash, ink)
// add the texture a polygon cannot; without them the drawing still reads.

import (
	"fmt"
	"math"
	"strings"
)

const (
	// LookKey is the render-node prop that carries the look ("watercolor").
	LookKey = "_look"
	// SeedKey is the render-node prop that carries the node's seed string
	// (its id), so each component gets its own brush marks.
	SeedKey = "_seed"

	// LookWatercolor selects the painted look.
	LookWatercolor = "watercolor"

	// FontPainted is the face used for labels in the watercolor look: a book
	// serif reads as hand-set lettering next to washes, where a UI sans does not.
	FontPainted = "'Iowan Old Style', 'Palatino Linotype', Palatino, 'Book Antiqua', 'URW Palladio L', Georgia, serif"
)

// Watercolor returns the painted variant of a theme.
func (t Theme) Watercolor() Theme {
	if t.Name == Dark.Name {
		return Theme{
			Name: t.Name, Look: LookWatercolor,
			Paper: "#1f1e1d", Ink: "#ebe5d9",
			Surface: "#1f1e1d", SurfaceAlt: "#2a2826", Inset: "#252321",
			Border: "#ebe5d9", Text: "#efe9dd", Muted: "#a8a194", Arrow: "#e3ddd0",
			CodeBg: "#0c0c0d", CodeFg: "#ebe5d9",
		}
	}
	return Theme{
		Name: t.Name, Look: LookWatercolor,
		Paper: "#f5efe3", Ink: "#2f3042",
		Surface: "#f5efe3", SurfaceAlt: "#e8e0d1", Inset: "#efe7d9",
		Border: "#2f3042", Text: "#2a2b3a", Muted: "#6f685d", Arrow: "#3a3b4d",
		CodeBg: "#273042", CodeFg: "#f3ede1",
	}
}

// IsWatercolor reports whether the theme paints instead of fills.
func (t Theme) IsWatercolor() bool { return t.Look == LookWatercolor }

// Pigments: each accent becomes the watercolor pigment closest to it, so the
// same component keeps its identity across looks, only softer and earthier.
var pigments = map[string]string{
	AccentBlue:   "#3d63ad", // cobalt
	AccentTeal:   "#1d847b", // viridian
	AccentAmber:  "#de961c", // gamboge
	AccentCyan:   "#2687b3", // cerulean
	AccentViolet: "#6a52a3", // dioxazine
	AccentPink:   "#c03e69", // quinacridone rose
	AccentGreen:  "#5a8a33", // sap green
	AccentSlate:  "#56637a", // payne's grey
}

// washFilters is how many differently seeded wash filters are defined.
const washFilters = 3

// denseWash is the density from which a wash uses the calm filter.
const denseWash = 0.5

// Named pigments used by the painted primitives.
const (
	pigmentPaynes  = "#5b6577"
	pigmentNeutral = "#857a6e" // neutral tint, warm: for shapes with no accent
	pigmentUmber   = "#8c7656"
	pigmentRose    = "#c8424b"
	pigmentYellow  = "#e2a020"
	pigmentSap     = "#5c9638"
)

// Pigment maps a colour to the pigment that paints it. Accent constants map
// to their pigment; anything else (a user's colour) is used as given.
func Pigment(c string) string {
	if p, ok := pigments[strings.ToLower(strings.TrimSpace(c))]; ok {
		return p
	}
	return c
}

// paintBlend picks how a wash meets the ground: multiply on paper (pigment
// only darkens), screen on a dark sheet (it lightens), and plain normal for a
// colour darker than a dark sheet (opaque, like gouache).
//
// On paper, a light colour is painted as body colour (opaque, like gouache):
// a transparent wash can only darken, so a pale card would vanish on a dark
// container under it.
func paintBlend(pigment string, t Theme) string {
	if t.Name != Dark.Name {
		if l, ok := luminance(pigment); ok && l > 0.45 {
			return "normal"
		}
		return "multiply"
	}
	lp, okP := luminance(pigment)
	lg, _ := luminance(t.Paper)
	if okP && lp < lg {
		return "normal"
	}
	return "screen"
}

// PaintedColor estimates the colour a wash of the given density leaves on the
// sheet. It is what text sitting on the wash must contrast with. The settled
// band, the rim and the layering add roughly half again to the nominal
// density, so the estimate does too.
func PaintedColor(pigment string, density float64, t Theme) string {
	density = Clamp(density*1.5, 0, 1)
	g, okG := rgbOf(t.Paper)
	p, okP := rgbOf(pigment)
	if !okG || !okP {
		return pigment
	}
	blend := paintBlend(pigment, t)
	var out [3]float64
	for i := range out {
		var full float64
		switch blend {
		case "multiply":
			full = g[i] * p[i]
		case "screen":
			full = 1 - (1-g[i])*(1-p[i])
		default:
			full = p[i]
		}
		out[i] = g[i] + (full-g[i])*density
	}
	return hexOf(out)
}

// WashSpec describes one painted area.
type WashSpec struct {
	Poly    []Pt    // closed outline, local px
	Pigment string  // colour of the paint
	Density float64 // 0..1: how much of the pigment reaches the core
	Seed    uint64
	Layers  int     // translucent layers; 0 means 4
	Wander  float64 // how far edges drift, relative to edge length; 0 means 0.09
	MaxOff  float64 // hard cap on drift in px; 0 means 5
	Settle  float64 // strength of the settled, darker lower band; 0 means none
	Pockets int     // loose blots of extra pigment inside the shape
	Filter  string  // filter id override; "none" paints without texture
	Blend   string  // blend override; "" picks one from the pigment and sheet
	Class   string
}

// Wash paints a translucent pigment area: a few deformed copies of the shape
// stacked at low opacity. Where all layers overlap the colour is full; at the
// edge only some do, which gives the soft, irregular watercolor boundary. Each
// layer's outline is also stroked faintly, and those strokes pile up into the
// darker rim pigment leaves where a wash dries. A couple of extra layers over
// the lower part of the shape are the pigment that settled as the paper dried
// at a slant: a graded wash with a soft tide line across it.
func Wash(s WashSpec, t Theme) string {
	if len(s.Poly) < 3 || s.Density <= 0 {
		return ""
	}
	layers := s.Layers
	if layers <= 0 {
		layers = 4
	}
	wander := s.Wander
	if wander <= 0 {
		wander = 0.09
	}
	maxOff := s.MaxOff
	if maxOff <= 0 {
		maxOff = 5
	}
	r := newRng(s.Seed)
	// Long straight edges need points along them to wander at all.
	per := polyLength(append(append([]Pt{}, s.Poly...), s.Poly[0]))
	poly := resample(s.Poly, Clamp(per/Clamp(per/8, 12, 36), 1, 40), true)
	base := deform(poly, 1, wander, maxOff, r)
	alpha := 1 - math.Pow(1-Clamp(s.Density, 0, 0.98), 1/float64(layers))

	// The filter's noise is sampled in the wash's own coordinates, so every
	// component would get the same texture. Paint each wash somewhere else in
	// the noise field and move it back with a transform.
	ox := math.Round(r.float() * 600)
	oy := math.Round(r.float() * 600)
	moved := func(poly []Pt) string {
		out := make([]Pt, len(poly))
		for i, p := range poly {
			out[i] = Pt{p.X + ox, p.Y + oy}
		}
		return relPathInt(out)
	}

	var b strings.Builder
	class := "nagare-wash"
	if s.Class != "" {
		class += " " + s.Class
	}
	filter := ""
	switch s.Filter {
	case "":
		filter = fmt.Sprintf(` filter="url(#nagare-wc-wash%d)"`, s.Seed%washFilters)
		if l, ok := luminance(s.Pigment); s.Density >= denseWash || (ok && l < 0.05) {
			// A heavy wash shows every bloom of the mottling; keep it calm.
			filter = ` filter="url(#nagare-wc-dense)"`
		}
	case "none":
	default:
		filter = fmt.Sprintf(` filter="url(#%s)"`, s.Filter)
	}
	fmt.Fprintf(&b, `<g class="%s"`, class)
	blend := s.Blend
	if blend == "" {
		blend = paintBlend(s.Pigment, t)
	}
	if blend != "normal" {
		fmt.Fprintf(&b, ` style="mix-blend-mode:%s"`, blend)
	}
	fmt.Fprintf(&b, `><g transform="translate(%s %s)" fill="%s" fill-opacity="%.3f" stroke="%s" stroke-opacity="%.3f" stroke-width="1" stroke-linejoin="round"%s>`,
		F(-ox), F(-oy), s.Pigment, alpha, s.Pigment, math.Min(1, alpha*1.5), filter)
	for i := 0; i < layers; i++ {
		poly := deform(base, 1, wander*0.9, maxOff*0.7, r)
		poly = jitter(poly, 0.35, r)
		fmt.Fprintf(&b, `<path d="%s"/>`, moved(poly))
	}
	// Pockets: loose blots inside the shape where more pigment gathered and
	// dried with a hard edge (a backrun).
	if s.Pockets > 0 {
		minX, minY, maxX, maxY := bounds(base)
		w, h := maxX-minX, maxY-minY
		for i := 0; i < s.Pockets; i++ {
			rx := w * (0.18 + r.float()*0.22)
			ry := h * (0.22 + r.float()*0.25)
			cx := minX + rx + r.float()*math.Max(0, w-2*rx)
			cy := minY + ry + r.float()*math.Max(0, h-2*ry)
			blob := deform(Ellipse(cx, cy, rx, ry, 10), 3, 0.22, math.Min(rx, ry)*0.4, r)
			fmt.Fprintf(&b, `<path d="%s" fill-opacity="%.3f" stroke="none"/>`, moved(blob), alpha*0.9)
		}
	}
	if settled := settle(base, r); settled != nil && s.Settle > 0 {
		poly := deform(settled, 1, wander, maxOff*0.8, r)
		fmt.Fprintf(&b, `<path d="%s" fill-opacity="%.3f" stroke="none"/>`, moved(poly), math.Min(1, alpha*s.Settle*1.8))
	}
	b.WriteString(`</g>`)
	b.WriteString(`</g>`)
	return b.String()
}

// settle returns the part of a closed outline below a wavering tide line,
// somewhere between 40% and 65% of the way down.
func settle(poly []Pt, r *rng) []Pt {
	minY, maxY := math.Inf(1), math.Inf(-1)
	minX, maxX := math.Inf(1), math.Inf(-1)
	for _, p := range poly {
		minY, maxY = math.Min(minY, p.Y), math.Max(maxY, p.Y)
		minX, maxX = math.Min(minX, p.X), math.Max(maxX, p.X)
	}
	h := maxY - minY
	if h < 8 {
		return nil
	}
	tide := minY + h*(0.4+r.float()*0.25)
	tilt := r.signed() * h * 0.18
	seed := r.next()
	line := func(x float64) float64 {
		k := (x - minX) / math.Max(1, maxX-minX)
		return tide + tilt*(k-0.5) + fbm1(seed, x/38)*h*0.06
	}
	out := make([]Pt, 0, len(poly))
	for _, p := range poly {
		if y := line(p.X); p.Y < y {
			p.Y = y
		}
		out = append(out, p)
	}
	return out
}

// InkSpec describes a pen line.
type InkSpec struct {
	Width    float64 // nominal nib width
	Color    string
	Seed     uint64
	Tremor   float64 // hand wobble in px; 0 means 0.7
	TaperIn  float64 // px over which the line swells from its start
	TaperOut float64 // px over which it lifts off at the end
	Opacity  float64 // 0 means opaque
	NoLift   bool    // InkLoop: never lift the pen
}

// InkStroke draws an open polyline as one pen stroke: a filled outline whose
// width follows a slow pressure curve and tapers where the nib lands and lifts.
func InkStroke(pts []Pt, s InkSpec) string {
	poly := inkOutline(pts, s)
	if len(poly) < 3 {
		return ""
	}
	op := ""
	if s.Opacity > 0 && s.Opacity < 1 {
		op = fmt.Sprintf(` fill-opacity="%s"`, F(s.Opacity))
	}
	return fmt.Sprintf(`<path d="%s" fill="%s"%s/>`, relPathD(poly), s.Color, op)
}

func inkOutline(pts []Pt, s InkSpec) []Pt {
	p := resample(pts, 4.2, false)
	n := len(p)
	if n < 2 {
		return nil
	}
	tremor := s.Tremor
	if tremor == 0 {
		tremor = 0.7
	}
	cum := make([]float64, n)
	for i := 1; i < n; i++ {
		cum[i] = cum[i-1] + math.Hypot(p[i].X-p[i-1].X, p[i].Y-p[i-1].Y)
	}
	total := cum[n-1]
	taper := func(d, over float64) float64 {
		if over <= 0 {
			return 1
		}
		k := Clamp(d/over, 0, 1)
		return 0.18 + 0.82*k*k*(3-2*k)
	}
	left := make([]Pt, n)
	right := make([]Pt, n)
	for i := range p {
		a := p[max(0, i-1)]
		c := p[min(n-1, i+1)]
		dx, dy := c.X-a.X, c.Y-a.Y
		l := math.Hypot(dx, dy)
		if l < 1e-9 {
			dx, dy, l = 1, 0, 1
		}
		nx, ny := -dy/l, dx/l
		wob := fbm1(s.Seed, cum[i]/48) * tremor
		press := 0.78 + 0.36*fbm1(s.Seed+11, cum[i]/90)
		hw := s.Width * press * taper(cum[i], s.TaperIn) * taper(total-cum[i], s.TaperOut) / 2
		cx, cy := p[i].X+nx*wob, p[i].Y+ny*wob
		left[i] = Pt{cx + nx*hw, cy + ny*hw}
		right[i] = Pt{cx - nx*hw, cy - ny*hw}
	}
	out := make([]Pt, 0, 2*n+2)
	out = append(out, left...)
	for i := n - 1; i >= 0; i-- {
		out = append(out, right[i])
	}
	return out
}

// InkLoop draws a closed outline as one stroke that starts somewhere on the
// shape, goes all the way round and runs a little past where it began, drifting
// off the line: a hand closing a box, not a plotter.
func InkLoop(poly []Pt, s InkSpec) string {
	p := resample(poly, 4.2, true)
	n := len(p)
	if n < 4 {
		return ""
	}
	r := newRng(s.Seed)
	start := int(r.float() * float64(n))
	overlap := 4 + int(r.float()*4)
	drift := 0.8 + r.float()*1.4
	// Centroid, to know which way is outward.
	var cx, cy float64
	for _, q := range p {
		cx += q.X
		cy += q.Y
	}
	cx /= float64(n)
	cy /= float64(n)
	walk := make([]Pt, 0, n+overlap)
	for i := 0; i <= n+overlap; i++ {
		q := p[(start+i)%n]
		if i > n-2 {
			k := float64(i-(n-2)) / float64(overlap+2)
			dx, dy := q.X-cx, q.Y-cy
			l := math.Max(1, math.Hypot(dx, dy))
			q = Pt{q.X + dx/l*drift*k, q.Y + dy/l*drift*k}
		}
		walk = append(walk, q)
	}
	if s.TaperIn == 0 {
		s.TaperIn = 10
	}
	if s.TaperOut == 0 {
		s.TaperOut = 16
	}
	// On a long enough outline the pen lifts once: a short lost stretch
	// somewhere in the middle of the walk, which the eye closes on its own.
	if polyLength(walk) > 220 && !s.NoLift {
		mid := int(float64(len(walk)) * (0.3 + r.float()*0.45))
		gap := 2 + int(r.float()*3)
		if mid-gap > 4 && mid+gap < len(walk)-4 {
			a, c := s, s
			a.TaperOut = 9
			c.TaperIn = 9
			c.Seed += 101
			return InkStroke(walk[:mid-gap+1], a) + InkStroke(walk[mid+gap:], c)
		}
	}
	return InkStroke(walk, s)
}

// InkArrow draws a connector: a pen line along the route with its corners
// eased, and a flicked arrowhead. Dashed lines are a run of short strokes.
func InkArrow(points []Pt, s InkSpec, dashed, head, tail bool) string {
	if len(points) < 2 {
		return ""
	}
	route := roundCorners(points, 12)
	var b strings.Builder
	b.WriteString(`<g class="nagare-ink-arrow">`)
	if dashed {
		b.WriteString(InkDashes(route, s, 9, 6.5))
	} else {
		ls := s
		ls.TaperIn = 8
		ls.TaperOut = 2
		b.WriteString(InkStroke(route, ls))
	}
	if head {
		b.WriteString(inkHead(route, s, false))
	}
	if tail {
		b.WriteString(inkHead(route, s, true))
	}
	b.WriteString(`</g>`)
	return b.String()
}

// InkDashes draws a polyline as a run of short pen strokes.
func InkDashes(pts []Pt, s InkSpec, dash, gap float64) string {
	if len(pts) < 2 {
		return ""
	}
	dense := resample(pts, 1, false)
	var b strings.Builder
	seg := []Pt{dense[0]}
	acc := 0.0
	on := true
	k := uint64(0)
	flush := func() {
		if len(seg) < 2 {
			// A dot: give the nib something to draw.
			p := seg[0]
			seg = append(seg, Pt{p.X + 0.6, p.Y})
		}
		ds := s
		ds.Seed += k
		ds.TaperIn, ds.TaperOut = math.Min(3, dash/2), math.Min(3, dash/2)
		b.WriteString(InkStroke(seg, ds))
		k++
	}
	for i := 1; i < len(dense); i++ {
		acc += math.Hypot(dense[i].X-dense[i-1].X, dense[i].Y-dense[i-1].Y)
		if on {
			seg = append(seg, dense[i])
			if acc >= dash || i == len(dense)-1 {
				flush()
				acc, on = 0, false
			}
		} else if acc >= gap {
			seg = []Pt{dense[i]}
			acc, on = 0, true
		}
	}
	return b.String()
}

// inkHead draws an arrowhead as two quick flicks of the pen into the tip.
func inkHead(route []Pt, s InkSpec, atStart bool) string {
	pts := route
	if atStart {
		pts = make([]Pt, len(route))
		for i := range route {
			pts[i] = route[len(route)-1-i]
		}
	}
	tip := pts[len(pts)-1]
	// Direction from a point ~10px back along the route.
	back := tip
	acc := 0.0
	for i := len(pts) - 2; i >= 0; i-- {
		acc += math.Hypot(pts[i+1].X-pts[i].X, pts[i+1].Y-pts[i].Y)
		back = pts[i]
		if acc >= 10 {
			break
		}
	}
	dx, dy := tip.X-back.X, tip.Y-back.Y
	l := math.Hypot(dx, dy)
	if l < 1e-6 {
		return ""
	}
	ux, uy := dx/l, dy/l
	r := newRng(s.Seed ^ 0xa11ce)
	length := 7.5 + s.Width*2.2
	var b strings.Builder
	for _, side := range []float64{1, -1} {
		ang := (0.46 + r.signed()*0.05) * side
		ca, sa := math.Cos(ang), math.Sin(ang)
		// Rotate the backwards direction by ang.
		bx := -ux*ca + uy*sa
		by := -ux*sa - uy*ca
		ln := length * (0.92 + r.float()*0.16)
		from := Pt{tip.X + bx*ln, tip.Y + by*ln}
		// A slight bow, so the flick is not a ruler line.
		mid := Pt{(from.X+tip.X)/2 - by*side*0.8, (from.Y+tip.Y)/2 + bx*side*0.8}
		hs := s
		hs.Seed += uint64(side+2) * 31
		hs.Width = s.Width * 1.15
		hs.TaperIn = ln * 0.7
		hs.TaperOut = 1
		hs.Tremor = 0.2
		b.WriteString(InkStroke([]Pt{from, mid, tip}, hs))
	}
	return b.String()
}

// WatercolorDefs returns the shared filters. Their ids are fixed and their
// content does not depend on the diagram, so several diagrams inlined in one
// page can share them safely.
func WatercolorDefs() string {
	var b strings.Builder
	b.WriteString(`<defs>`)
	// Paper: big soft clouds in the sheet plus cold-press tooth, as alpha over
	// the fill colour of the element it is applied to.
	b.WriteString(`<filter id="nagare-wc-paper" x="0" y="0" width="100%" height="100%" color-interpolation-filters="sRGB">` +
		`<feTurbulence type="fractalNoise" baseFrequency="0.0045" numOctaves="4" seed="11" result="m"/>` +
		`<feColorMatrix in="m" type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1.1 0 0 0 -0.42" result="ma"/>` +
		`<feTurbulence type="fractalNoise" baseFrequency="0.7" numOctaves="2" seed="5" result="t"/>` +
		`<feColorMatrix in="t" type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1.4 0 0 0 -0.55" result="ta"/>` +
		`<feMerge result="a"><feMergeNode in="ma"/><feMergeNode in="ta"/></feMerge>` +
		`<feComposite in="SourceGraphic" in2="a" operator="in"/>` +
		`</filter>`)
	// Wash: the shape's edge wanders (displacement), the pigment is uneven
	// (large soft mottling, so some areas bloom lighter), granulates into
	// fine dark specks, and dries darker at the rim (shape minus its eroded
	// self). Several seeds so neighbouring washes do not share a texture.
	for i := 0; i < washFilters; i++ {
		fmt.Fprintf(&b, `<filter id="nagare-wc-wash%d" x="-10%%" y="-20%%" width="120%%" height="140%%" color-interpolation-filters="sRGB">`+
			`<feTurbulence type="fractalNoise" baseFrequency="0.022" numOctaves="3" seed="%d" result="warp"/>`+
			`<feDisplacementMap in="SourceGraphic" in2="warp" scale="6" xChannelSelector="R" yChannelSelector="G" result="shape"/>`+
			`<feTurbulence type="fractalNoise" baseFrequency="0.009" numOctaves="2" seed="%d" result="mottle"/>`+
			`<feColorMatrix in="mottle" type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 2.5 0 0 0 -0.5" result="mA"/>`+
			`<feComposite in="shape" in2="mA" operator="in" result="body"/>`+
			`<feTurbulence type="fractalNoise" baseFrequency="1.3" numOctaves="1" seed="%d" result="grain"/>`+
			`<feColorMatrix in="grain" type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 3.2 0 0 0 -1.75" result="gA"/>`+
			`<feComposite in="shape" in2="gA" operator="in" result="specks"/>`+
			`<feMorphology in="shape" operator="erode" radius="1.4" result="er"/>`+
			`<feComposite in="shape" in2="er" operator="out" result="rim"/>`+
			`<feGaussianBlur in="rim" stdDeviation="0.6" result="rimS"/>`+
			`<feMerge><feMergeNode in="body"/><feMergeNode in="specks"/><feMergeNode in="rimS"/><feMergeNode in="rimS"/></feMerge>`+
			`</filter>`, i, 4+i*13, 8+i*7, 2+i*5)
	}
	// Dense: the same edge and rim, but only a gentle unevenness and no
	// specks, for heavy washes such as a terminal's body.
	b.WriteString(`<filter id="nagare-wc-dense" x="-10%" y="-20%" width="120%" height="140%" color-interpolation-filters="sRGB">` +
		`<feTurbulence type="fractalNoise" baseFrequency="0.022" numOctaves="3" seed="31" result="warp"/>` +
		`<feDisplacementMap in="SourceGraphic" in2="warp" scale="5" xChannelSelector="R" yChannelSelector="G" result="shape"/>` +
		`<feTurbulence type="fractalNoise" baseFrequency="0.012" numOctaves="2" seed="37" result="mottle"/>` +
		`<feColorMatrix in="mottle" type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0.45 0 0 0 0.74" result="mA"/>` +
		`<feComposite in="shape" in2="mA" operator="in" result="body"/>` +
		`<feMorphology in="shape" operator="erode" radius="1.4" result="er"/>` +
		`<feComposite in="shape" in2="er" operator="out" result="rim"/>` +
		`<feGaussianBlur in="rim" stdDeviation="0.6" result="rimS"/>` +
		`<feMerge><feMergeNode in="body"/><feMergeNode in="rimS"/></feMerge>` +
		`</filter>`)
	// Soft: wet-in-wet. The edge wanders and melts, the pigment mottles, and
	// there is no dried rim: for glazes that should fade rather than stop.
	b.WriteString(`<filter id="nagare-wc-soft" x="-10%" y="-20%" width="120%" height="140%" color-interpolation-filters="sRGB">` +
		`<feTurbulence type="fractalNoise" baseFrequency="0.018" numOctaves="3" seed="41" result="warp"/>` +
		`<feDisplacementMap in="SourceGraphic" in2="warp" scale="9" xChannelSelector="R" yChannelSelector="G" result="shape"/>` +
		`<feGaussianBlur in="shape" stdDeviation="2.4" result="soft"/>` +
		`<feTurbulence type="fractalNoise" baseFrequency="0.01" numOctaves="2" seed="43" result="mottle"/>` +
		`<feColorMatrix in="mottle" type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1.8 0 0 0 -0.1" result="mA"/>` +
		`<feComposite in="soft" in2="mA" operator="in"/>` +
		`</filter>`)
	// Ink: a hint of the line soaking into the fibres.
	b.WriteString(`<filter id="nagare-wc-ink" x="-5%" y="-5%" width="110%" height="110%">` +
		`<feGaussianBlur in="SourceGraphic" stdDeviation="0.7" result="s"/>` +
		`<feComponentTransfer in="s" result="h"><feFuncA type="linear" slope="0.45"/></feComponentTransfer>` +
		`<feMerge><feMergeNode in="h"/><feMergeNode in="SourceGraphic"/></feMerge>` +
		`</filter>`)
	b.WriteString(`<radialGradient id="nagare-wc-vignette" cx="50%" cy="48%" r="75%">` +
		`<stop offset="0.55" stop-color="#000" stop-opacity="0"/>` +
		`<stop offset="1" stop-color="#000" stop-opacity="0.07"/>` +
		`</radialGradient>`)
	b.WriteString(`</defs>`)
	return b.String()
}

// Paper draws the sheet: base tone, texture and a faint darkening toward the
// deckle.
func Paper(w, h float64, t Theme) string {
	texture := "#6b5537"
	opacity := "0.13"
	if t.Name == Dark.Name {
		texture = "#d8cfbf"
		opacity = "0.06"
	}
	return fmt.Sprintf(`<g class="nagare-paper"><rect width="%s" height="%s" fill="%s"/>`+
		`<rect width="%s" height="%s" fill="%s" opacity="%s" filter="url(#nagare-wc-paper)"/>`+
		`<rect width="%s" height="%s" fill="url(#nagare-wc-vignette)"/></g>`,
		F(w), F(h), t.Paper, F(w), F(h), texture, opacity, F(w), F(h))
}

// WatercolorCSS returns the label typography for the painted look, scoped to
// painted diagrams so it is safe to inline next to other SVGs.
func WatercolorCSS() string {
	return `<style>svg.nagare-watercolor text:not([font-family^="ui-monospace"]){font-family:` + FontPainted + `}</style>`
}

// relPathD writes a closed polygon with relative coordinates, which is
// roughly a third smaller than absolute ones. Deltas are taken between
// rounded points so error never accumulates.
func relPathD(poly []Pt) string {
	if len(poly) == 0 {
		return ""
	}
	var b strings.Builder
	b.Grow(len(poly) * 8)
	r10 := func(v float64) int64 { return int64(math.Round(v * 10)) }
	px, py := r10(poly[0].X), r10(poly[0].Y)
	b.WriteByte('M')
	b.WriteString(tenths(px))
	b.WriteByte(' ')
	b.WriteString(tenths(py))
	b.WriteByte('l')
	for i := 1; i < len(poly); i++ {
		x, y := r10(poly[i].X), r10(poly[i].Y)
		dx, dy := x-px, y-py
		if i > 1 {
			if dx >= 0 {
				b.WriteByte(' ')
			}
		}
		b.WriteString(tenths(dx))
		if dy >= 0 {
			b.WriteByte(' ')
		}
		b.WriteString(tenths(dy))
		px, py = x, y
	}
	b.WriteByte('z')
	return b.String()
}

// relPathInt is relPathD at whole-pixel precision, for wash layers whose
// edges the filter displaces anyway.
func relPathInt(poly []Pt) string {
	if len(poly) == 0 {
		return ""
	}
	var b strings.Builder
	b.Grow(len(poly) * 6)
	px, py := int64(math.Round(poly[0].X)), int64(math.Round(poly[0].Y))
	fmt.Fprintf(&b, "M%d %dl", px, py)
	first := true
	for _, p := range poly[1:] {
		x, y := int64(math.Round(p.X)), int64(math.Round(p.Y))
		dx, dy := x-px, y-py
		if dx == 0 && dy == 0 {
			continue
		}
		if !first && dx >= 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%d", dx)
		if dy >= 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%d", dy)
		px, py, first = x, y, false
	}
	b.WriteByte('z')
	return b.String()
}

// tenths formats an integer count of tenths as a short decimal: 12 → "1.2",
// -3 → "-.3", 40 → "4".
func tenths(v int64) string {
	neg := v < 0
	if neg {
		v = -v
	}
	whole, frac := v/10, v%10
	var s string
	switch {
	case frac == 0:
		s = fmt.Sprintf("%d", whole)
	case whole == 0:
		s = fmt.Sprintf(".%d", frac)
	default:
		s = fmt.Sprintf("%d.%d", whole, frac)
	}
	if neg {
		return "-" + s
	}
	return s
}

func bounds(poly []Pt) (minX, minY, maxX, maxY float64) {
	minX, minY = math.Inf(1), math.Inf(1)
	maxX, maxY = math.Inf(-1), math.Inf(-1)
	for _, p := range poly {
		minX, maxX = math.Min(minX, p.X), math.Max(maxX, p.X)
		minY, maxY = math.Min(minY, p.Y), math.Max(maxY, p.Y)
	}
	return
}
