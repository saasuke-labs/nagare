// Package style is the shared visual system for diagram components: theme
// tokens, a small accent palette, line icons, text fitting, and the two
// primitives every component is drawn with (Card and Window).
package style

import (
	"fmt"
	"html"
	"math"
	"strings"
)

const (
	FontSans = "ui-sans-serif, system-ui, -apple-system, 'Segoe UI', Helvetica, Arial, sans-serif"
	FontMono = "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"

	// ThemeKey is the render-node prop that carries the theme name.
	ThemeKey = "_theme"

	// StrokeWidth is the outline weight shared by every component.
	StrokeWidth = 1.5
)

// Accent colours. Each reads on both the light and the dark surface.
const (
	AccentBlue   = "#3b82f6"
	AccentTeal   = "#14b8a6"
	AccentAmber  = "#f59e0b"
	AccentCyan   = "#06b6d4"
	AccentViolet = "#8b5cf6"
	AccentPink   = "#ec4899"
	AccentGreen  = "#22c55e"
	AccentSlate  = "#64748b"
)

// Theme holds the colours shared by every component.
type Theme struct {
	Name       string
	Surface    string // card and page fill
	SurfaceAlt string // window chrome
	Inset      string // container content area
	Border     string
	Text       string
	Muted      string
	Arrow      string
	CodeBg     string // terminal body
	CodeFg     string

	// Look is "" for the flat look or LookWatercolor for the painted one.
	Look  string
	Paper string // sheet colour (painted look)
	Ink   string // pen colour (painted look)
	// Seed varies brush marks per node. The renderer sets it from the id.
	Seed uint64
}

var (
	Light = Theme{
		Name: "light", Surface: "#ffffff", SurfaceAlt: "#eef0f3", Inset: "#f7f8fa",
		Border: "#cfd4dc", Text: "#1d2433", Muted: "#667085", Arrow: "#475467",
		CodeBg: "#1f2430", CodeFg: "#e6e9ef",
	}
	Dark = Theme{
		Name: "dark", Surface: "#1c2028", SurfaceAlt: "#262b35", Inset: "#15181e",
		Border: "#3b4350", Text: "#e6e9ef", Muted: "#98a2b3", Arrow: "#98a2b3",
		CodeBg: "#0d1015", CodeFg: "#e6e9ef",
	}
)

// ThemeByName returns the named theme; anything but "dark" is the light theme.
func ThemeByName(name string) Theme {
	if strings.EqualFold(strings.TrimSpace(name), "dark") {
		return Dark
	}
	return Light
}

// ThemeFromProps reads the theme a render node was given, including the
// look and the node's seed.
func ThemeFromProps(props map[string]any) Theme {
	name, _ := props[ThemeKey].(string)
	t := ThemeByName(name)
	if look, _ := props[LookKey].(string); strings.EqualFold(strings.TrimSpace(look), LookWatercolor) {
		t = t.Watercolor()
	}
	if seed, ok := props[SeedKey].(string); ok {
		t.Seed = SeedFor(seed)
	}
	return t
}

// CopyThemeProps copies the theme, look and seed props from src to dst, for
// composites that draw their parts through other components.
func CopyThemeProps(dst, src map[string]any, seedSuffix string) {
	for _, k := range []string{ThemeKey, LookKey} {
		if v, ok := src[k]; ok {
			dst[k] = v
		}
	}
	if seed, ok := src[SeedKey].(string); ok {
		dst[SeedKey] = seed + seedSuffix
	}
}

// Or returns override when it is set, otherwise fallback.
func Or(override, fallback string) string {
	if strings.TrimSpace(override) != "" {
		return override
	}
	return fallback
}

// Clamp limits v to [lo, hi].
func Clamp(v, lo, hi float64) float64 {
	return math.Max(lo, math.Min(hi, v))
}

// F formats a coordinate compactly.
func F(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	if s == "" || s == "-0" {
		return "0"
	}
	return s
}

// Esc escapes text for use in SVG content or attributes.
func Esc(s string) string { return html.EscapeString(s) }

// Fit truncates text with an ellipsis so it fits width at the given font size.
// Widths are estimated, since SVG gives no text metrics at render time.
func Fit(text string, width, fontSize, charRatio float64) string {
	if fontSize <= 0 {
		return text
	}
	maxChars := int(width / (fontSize * charRatio))
	runes := []rune(text)
	if len(runes) <= maxChars {
		return text
	}
	if maxChars <= 1 {
		return ""
	}
	return strings.TrimRight(string(runes[:maxChars-1]), " ") + "…"
}

const (
	sansRatio = 0.56
	monoRatio = 0.62
)

// TextOpts describes one line of text.
type TextOpts struct {
	X, Y    float64
	Size    float64
	Color   string
	Opacity float64 // 0 means fully opaque
	Anchor  string  // start | middle | end
	Weight  string
	Mono    bool
	Width   float64 // when > 0 the text is truncated to fit
}

// Text renders one line of text, vertically centred on Y.
func Text(s string, o TextOpts) string {
	ratio, family := sansRatio, FontSans
	if o.Mono {
		ratio, family = monoRatio, FontMono
	}
	if o.Width > 0 {
		s = Fit(s, o.Width, o.Size, ratio)
	}
	if s == "" {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<text x="%s" y="%s" font-family="%s" font-size="%s" fill="%s" dominant-baseline="central"`,
		F(o.X), F(o.Y), family, F(o.Size), o.Color)
	if o.Anchor != "" && o.Anchor != "start" {
		fmt.Fprintf(&b, ` text-anchor="%s"`, o.Anchor)
	}
	if o.Weight != "" {
		fmt.Fprintf(&b, ` font-weight="%s"`, o.Weight)
	}
	if o.Opacity > 0 && o.Opacity < 1 {
		fmt.Fprintf(&b, ` fill-opacity="%s"`, F(o.Opacity))
	}
	fmt.Fprintf(&b, `>%s</text>`, Esc(s))
	return b.String()
}

// Radius is the shared corner radius for a box of the given size.
func Radius(w, h float64) float64 {
	return Clamp(math.Min(w, h)*0.12, 3, 10)
}

// Icon names.
const (
	IconServer  = "server"
	IconQueue   = "queue"
	IconGlobe   = "globe"
	IconGateway = "gateway"
	IconClock   = "clock"
	IconPackage = "package"
	IconFile    = "file"
)

// Line icons on a 24x24 grid, drawn with a round 1.75 stroke.
var icons = map[string]string{
	IconServer:  `<rect x="3" y="4" width="18" height="7" rx="2"/><rect x="3" y="13" width="18" height="7" rx="2"/><path d="M7 7.5h.01M7 16.5h.01M11 7.5h6M11 16.5h6"/>`,
	IconQueue:   `<rect x="3" y="6" width="4.5" height="12" rx="1.2"/><rect x="9.75" y="6" width="4.5" height="12" rx="1.2"/><rect x="16.500" y="6" width="4.5" height="12" rx="1.2"/>`,
	IconGlobe:   `<circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3c2.6 2.5 3.900 5.500 3.900 9s-1.300 6.500-3.900 9c-2.600-2.500-3.900-5.500-3.900-9s1.300-6.500 3.900-9z"/>`,
	IconGateway: `<circle cx="5" cy="12" r="2"/><circle cx="19" cy="6" r="2"/><circle cx="19" cy="18" r="2"/><path d="M7 12h3.500c2 0 2.500-6 4.500-6h2M10.500 12c2 0 2.500 6 4.500 6h2"/>`,
	IconClock:   `<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3.200 2"/>`,
	IconPackage: `<path d="M12 3l8 4.500v9L12 21l-8-4.500v-9L12 3z"/><path d="M4 7.500l8 4.500 8-4.500M12 12v9"/>`,
	IconFile:    `<path d="M14 3H7.500A1.500 1.500 0 0 0 6 4.500v15A1.500 1.500 0 0 0 7.500 21h9a1.500 1.500 0 0 0 1.500-1.500V7l-4-4z"/><path d="M14 3v4h4M9.500 12.500h5M9.500 16h5"/>`,
}

// iconChip draws an icon inside a tinted rounded square of the given size.
func iconChip(name, accent string, x, y, size float64, s CardSpec) string {
	if s.Theme.IsWatercolor() {
		return iconChipPainted(name, accent, x, y, size, s.surface, s.Theme)
	}
	body, ok := icons[name]
	if !ok {
		return ""
	}
	pad := size * 0.2
	scale := (size - 2*pad) / 24
	return fmt.Sprintf(`<rect x="%s" y="%s" width="%s" height="%s" rx="%s" fill="%s" fill-opacity="0.14"/>`+
		`<g transform="translate(%s,%s) scale(%s)" fill="none" stroke="%s" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">%s</g>`,
		F(x), F(y), F(size), F(size), F(size*0.25), accent,
		F(x+pad), F(y+pad), fmt.Sprintf("%.4f", scale), accent, body)
}

// Box draws the shared outlined rounded rectangle.
func Box(w, h float64, fill, stroke string) string {
	r := Radius(w, h)
	half := StrokeWidth / 2
	return fmt.Sprintf(`<rect x="%s" y="%s" width="%s" height="%s" rx="%s" fill="%s" stroke="%s" stroke-width="%s"/>`,
		F(half), F(half), F(math.Max(0, w-StrokeWidth)), F(math.Max(0, h-StrokeWidth)), F(r), fill, stroke, F(StrokeWidth))
}

// CardSpec describes a card: the shape every leaf component shares.
type CardSpec struct {
	X, Y, W, H float64
	Class      string
	Title      string
	Subtitle   string
	Meta       string
	Icon       string
	Accent     string
	Bg, Fg     string // explicit overrides; empty means use the theme
	Theme      Theme

	surface string // painted look: the colour the card's content sits on
}

// Card renders a card. Wide, short boxes lay out as a row (icon left, text
// right); everything else stacks the icon above centred text. Lines that do
// not fit are dropped, never overlapped.
func Card(s CardSpec) string {
	if s.Theme.IsWatercolor() {
		return cardPainted(s)
	}
	t := s.Theme
	fill := Or(s.Bg, t.Surface)
	titleColor, mutedColor, mutedOpacity := TextColors(s.Fg, fill, t)
	accent := Or(s.Accent, t.Muted)

	var b strings.Builder
	fmt.Fprintf(&b, `<g class="nagare-card %s" transform="translate(%s, %s)">`, s.Class, F(s.X), F(s.Y))
	b.WriteString(Box(s.W, s.H, fill, t.Border))

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

func cardRow(s CardSpec, pad float64, lines []string, accent, titleColor, mutedColor string, mutedOpacity float64) string {
	var b strings.Builder
	titleSize := Clamp(s.H*0.26, 9, 15)
	subSize := Clamp(titleSize*0.8, 8, 12)

	textX := pad
	hasIcon := s.Icon != "" && s.W >= 70 && s.H >= 22
	if hasIcon {
		chip := Clamp(s.H-2*pad, 14, 36)
		b.WriteString(iconChip(s.Icon, accent, pad, (s.H-chip)/2, chip, s))
		textX = pad + chip + pad*0.8
	}
	textW := s.W - textX - pad

	anchor, x := "start", textX
	if !hasIcon {
		anchor, x = "middle", s.W/2
		textW = s.W - 2*pad
	}

	sub := strings.Join(lines, " · ")
	if sub != "" && s.H >= titleSize+subSize+10 {
		gap := subSize * 0.35
		top := (s.H - (titleSize + gap + subSize)) / 2
		b.WriteString(Text(s.Title, TextOpts{X: x, Y: top + titleSize/2, Size: titleSize, Color: titleColor, Weight: "600", Anchor: anchor, Width: textW}))
		b.WriteString(Text(sub, TextOpts{X: x, Y: top + titleSize + gap + subSize/2, Size: subSize, Color: mutedColor, Opacity: mutedOpacity, Anchor: anchor, Width: textW}))
	} else {
		b.WriteString(Text(s.Title, TextOpts{X: x, Y: s.H / 2, Size: titleSize, Color: titleColor, Weight: "600", Anchor: anchor, Width: textW}))
	}
	return b.String()
}

func cardStack(s CardSpec, pad float64, lines []string, accent, titleColor, mutedColor string, mutedOpacity float64) string {
	var b strings.Builder
	base := math.Min(s.W, s.H)
	titleSize := Clamp(base*0.1, 11, 17)
	subSize := Clamp(titleSize*0.78, 9, 12.5)
	gap := subSize * 0.45
	chip := 0.0
	if s.Icon != "" {
		chip = Clamp(base*0.26, 22, 44)
	}
	chipGap := titleSize * 0.7

	avail := s.H - 2*pad
	height := func(n int) float64 {
		h := titleSize + float64(n)*(gap+subSize)
		if chip > 0 {
			h += chip + chipGap
		}
		return h
	}
	// Drop detail lines, then the icon, until the block fits.
	for len(lines) > 0 && height(len(lines)) > avail {
		lines = lines[:len(lines)-1]
	}
	if height(len(lines)) > avail {
		chip = 0
	}

	y := (s.H - height(len(lines))) / 2
	cx := s.W / 2
	textW := s.W - 2*pad
	if chip > 0 {
		b.WriteString(iconChip(s.Icon, accent, cx-chip/2, y, chip, s))
		y += chip + chipGap
	}
	b.WriteString(Text(s.Title, TextOpts{X: cx, Y: y + titleSize/2, Size: titleSize, Color: titleColor, Weight: "600", Anchor: "middle", Width: textW}))
	y += titleSize
	for _, l := range lines {
		y += gap
		b.WriteString(Text(l, TextOpts{X: cx, Y: y + subSize/2, Size: subSize, Color: mutedColor, Opacity: mutedOpacity, Anchor: "middle", Width: textW}))
		y += subSize
	}
	return b.String()
}

// Traffic-light colours for window controls.
var controlColors = [3]string{"#ff5f57", "#febc2e", "#28c840"}

// WindowSpec describes a window: the shape browser, VM and terminal share.
type WindowSpec struct {
	X, Y, W, H float64
	Class      string
	BarH       float64 // height of the title bar
	// Content area rectangle, relative to the window.
	CX, CY, CW, CH float64
	Title          string
	TitlePill      bool // draw the title as an address pill
	TitleMono      bool
	Bg, Fg         string // overrides for chrome fill and title colour
	ContentFill    string
	Theme          Theme
}

// Window renders the chrome and content area. The returned SVG leaves the
// group open so callers can add content; close it with WindowEnd.
func Window(s WindowSpec) string {
	if s.Theme.IsWatercolor() {
		return windowPainted(s)
	}
	t := s.Theme
	chrome := Or(s.Bg, t.SurfaceAlt)
	titleFill := chrome
	if s.TitlePill {
		titleFill = t.Surface
	}
	titleColor, titleOpacity := Legible(s.Fg, titleFill, t.Muted), 0.0
	if titleColor != t.Muted {
		titleOpacity = 0.85
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<g class="nagare-window %s" transform="translate(%s, %s)">`, s.Class, F(s.X), F(s.Y))
	b.WriteString(Box(s.W, s.H, chrome, t.Border))

	contentR := Clamp(Radius(s.W, s.H)-3, 2, 7)
	fmt.Fprintf(&b, `<rect x="%s" y="%s" width="%s" height="%s" rx="%s" fill="%s" stroke="%s" stroke-width="1"/>`,
		F(s.CX), F(s.CY), F(s.CW), F(s.CH), F(contentR), s.ContentFill, t.Border)

	// Controls
	r := Clamp(s.BarH*0.14, 2, 5.5)
	step := r * 3.1
	cx := Clamp(s.BarH*0.42, r+4, 22)
	cy := s.BarH / 2
	for i, c := range controlColors {
		fmt.Fprintf(&b, `<circle cx="%s" cy="%s" r="%s" fill="%s"/>`, F(cx+float64(i)*step), F(cy), F(r), c)
	}
	controlsEnd := cx + 2*step + r

	// Title
	size := Clamp(s.BarH*0.36, 8, 13)
	left := controlsEnd + size
	right := s.W - left
	if right-left > size*3 && s.Title != "" {
		if s.TitlePill {
			ph := Clamp(s.BarH*0.56, size+4, 26)
			fmt.Fprintf(&b, `<rect x="%s" y="%s" width="%s" height="%s" rx="%s" fill="%s" stroke="%s" stroke-width="1"/>`,
				F(left), F(cy-ph/2), F(right-left), F(ph), F(ph/2), t.Surface, t.Border)
			b.WriteString(Text(s.Title, TextOpts{X: left + ph*0.55, Y: cy, Size: size, Color: titleColor, Opacity: titleOpacity, Mono: s.TitleMono, Width: right - left - ph*1.1}))
		} else {
			b.WriteString(Text(s.Title, TextOpts{X: s.W / 2, Y: cy, Size: size, Color: titleColor, Opacity: titleOpacity, Anchor: "middle", Weight: "600", Mono: s.TitleMono, Width: right - left}))
		}
	}
	return b.String()
}

// WindowEnd closes the group opened by Window.
const WindowEnd = `</g>`
