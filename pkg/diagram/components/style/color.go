package style

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// luminance returns the WCAG relative luminance of a #rgb / #rrggbb colour.
// ok is false for anything else (named colours, rgb(), gradients...).
func luminance(c string) (l float64, ok bool) {
	c = strings.TrimPrefix(strings.TrimSpace(c), "#")
	if len(c) == 3 {
		c = string([]byte{c[0], c[0], c[1], c[1], c[2], c[2]})
	}
	if len(c) != 6 {
		return 0, false
	}
	v, err := strconv.ParseUint(c, 16, 32)
	if err != nil {
		return 0, false
	}
	lin := func(u uint64) float64 {
		s := float64(u&0xff) / 255
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(v>>16) + 0.7152*lin(v>>8) + 0.0722*lin(v), true
}

// Contrast returns the WCAG contrast ratio between two colours, or 21 when
// either cannot be parsed (so unknown colours are trusted as given).
func Contrast(a, b string) float64 {
	la, okA := luminance(a)
	lb, okB := luminance(b)
	if !okA || !okB {
		return 21
	}
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// OnColor returns a dark or light text colour that reads well on fill.
func OnColor(fill string) string {
	if Contrast(Light.Text, fill) >= Contrast(Dark.Text, fill) {
		return Light.Text
	}
	return Dark.Text
}

// Legible returns preferred when it is set and readable on fill, otherwise
// fallback when that is readable, otherwise a plain dark/light text colour.
func Legible(preferred, fill, fallback string) string {
	const min = 3.0
	if strings.TrimSpace(preferred) != "" && Contrast(preferred, fill) >= min {
		return preferred
	}
	if Contrast(fallback, fill) >= min {
		return fallback
	}
	return OnColor(fill)
}

// TextColors picks the title and secondary text colours for text drawn on
// fill. An fg override wins when it is readable; a bg override without a
// usable fg gets a plain contrasting colour; otherwise the theme applies.
func TextColors(fg, fill string, t Theme) (title, muted string, mutedOpacity float64) {
	title = Legible(fg, fill, t.Text)
	if title == t.Text && Contrast(t.Muted, fill) >= 3 {
		return title, t.Muted, 0
	}
	return title, title, 0.7
}

// rgbOf parses #rgb / #rrggbb into channels in [0, 1].
func rgbOf(c string) (rgb [3]float64, ok bool) {
	c = strings.TrimPrefix(strings.TrimSpace(c), "#")
	if len(c) == 3 {
		c = string([]byte{c[0], c[0], c[1], c[1], c[2], c[2]})
	}
	if len(c) != 6 {
		return rgb, false
	}
	v, err := strconv.ParseUint(c, 16, 32)
	if err != nil {
		return rgb, false
	}
	return [3]float64{float64(v>>16&0xff) / 255, float64(v>>8&0xff) / 255, float64(v&0xff) / 255}, true
}

// hexOf formats channels in [0, 1] as #rrggbb.
func hexOf(rgb [3]float64) string {
	b := func(v float64) int { return int(math.Round(Clamp(v, 0, 1) * 255)) }
	return fmt.Sprintf("#%02x%02x%02x", b(rgb[0]), b(rgb[1]), b(rgb[2]))
}

// Shade mixes a colour toward black (amount > 0) or white (amount < 0).
// Unparseable colours are returned unchanged.
func Shade(c string, amount float64) string {
	rgb, ok := rgbOf(c)
	if !ok {
		return c
	}
	for i := range rgb {
		if amount >= 0 {
			rgb[i] *= 1 - amount
		} else {
			rgb[i] += (1 - rgb[i]) * -amount
		}
	}
	return hexOf(rgb)
}
