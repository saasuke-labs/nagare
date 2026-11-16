package style

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// parseRel decodes the "M x y l dx dy ... z" paths the painted look writes.
func parseRel(t *testing.T, d string) []Pt {
	t.Helper()
	if !strings.HasPrefix(d, "M") || !strings.HasSuffix(d, "z") {
		t.Fatalf("unexpected path %q", d)
	}
	nums := regexp.MustCompile(`-?(\d+\.?\d*|\.\d+)`).FindAllString(d, -1)
	if len(nums)%2 != 0 {
		t.Fatalf("odd number count in %q", d)
	}
	var pts []Pt
	var x, y float64
	for i := 0; i < len(nums); i += 2 {
		a, _ := strconv.ParseFloat(nums[i], 64)
		b, _ := strconv.ParseFloat(nums[i+1], 64)
		if i == 0 {
			x, y = a, b
		} else {
			x, y = x+a, y+b
		}
		pts = append(pts, Pt{x, y})
	}
	return pts
}

func TestRelativePathsRoundTrip(t *testing.T) {
	poly := []Pt{{10.04, 20.06}, {12.5, 19.94}, {12.55, 25}, {-3.3, 0.04}, {0, 0}, {100.26, -7.71}}
	got := parseRel(t, relPathD(poly))
	if len(got) != len(poly) {
		t.Fatalf("got %d points, want %d", len(got), len(poly))
	}
	for i := range poly {
		if math.Abs(got[i].X-poly[i].X) > 0.051 || math.Abs(got[i].Y-poly[i].Y) > 0.051 {
			t.Errorf("point %d: got %v, want %v (±0.05)", i, got[i], poly[i])
		}
	}

	got = parseRel(t, relPathInt(poly))
	for _, p := range got {
		if p.X != math.Round(p.X) || p.Y != math.Round(p.Y) {
			t.Errorf("expected whole pixels, got %v", p)
		}
	}
}

func TestTenths(t *testing.T) {
	for in, want := range map[int64]string{0: "0", 12: "1.2", -3: "-.3", 40: "4", -125: "-12.5", 5: ".5"} {
		if got := tenths(in); got != want {
			t.Errorf("tenths(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestWashIsSeededAndBounded(t *testing.T) {
	theme := Light.Watercolor()
	spec := WashSpec{Poly: RoundedRect(0, 0, 200, 60, 8, 24), Pigment: "#3d63ad", Density: 0.3, Seed: 42, MaxOff: 4, Settle: 0.6}
	a, b := Wash(spec, theme), Wash(spec, theme)
	if a != b {
		t.Fatal("expected the same seed to paint the same wash")
	}
	spec.Seed = 43
	if Wash(spec, theme) == a {
		t.Fatal("expected a different seed to paint a different wash")
	}

	// The wash is drawn offset into the noise field and moved back; undo
	// that and check every vertex stays near the shape.
	m := regexp.MustCompile(`translate\((-?[\d.]+) (-?[\d.]+)\)`).FindStringSubmatch(a)
	if m == nil {
		t.Fatal("expected the wash to be translated back from its noise offset")
	}
	ox, _ := strconv.ParseFloat(m[1], 64)
	oy, _ := strconv.ParseFloat(m[2], 64)
	for _, d := range regexp.MustCompile(` d="([^"]+)"`).FindAllStringSubmatch(a, -1) {
		for _, p := range parseRel(t, d[1]) {
			x, y := p.X+ox, p.Y+oy
			if x < -12 || x > 212 || y < -12 || y > 72 {
				t.Fatalf("wash vertex %v strays too far from a 200x60 shape", Pt{x, y})
			}
		}
	}
}

func TestPaintedColorAndBlend(t *testing.T) {
	light := Light.Watercolor()
	if got := paintBlend("#3d63ad", light); got != "multiply" {
		t.Errorf("pigment on paper: got %s, want multiply", got)
	}
	if got := paintBlend("#fef3c7", light); got != "normal" {
		t.Errorf("pale colour on paper: got %s, want normal (body colour)", got)
	}
	dark := Dark.Watercolor()
	if got := paintBlend("#3d63ad", dark); got != "screen" {
		t.Errorf("pigment on a dark sheet: got %s, want screen", got)
	}
	if got := paintBlend("#0c0c0d", dark); got != "normal" {
		t.Errorf("near-black on a dark sheet: got %s, want normal", got)
	}

	if got := PaintedColor("#3d63ad", 0, light); got != light.Paper {
		t.Errorf("no pigment should leave the paper: got %s", got)
	}
	l0, _ := luminance(light.Paper)
	l1, _ := luminance(PaintedColor("#3d63ad", 0.5, light))
	if l1 >= l0 {
		t.Error("expected a wash to darken paper")
	}
	if got := PaintedColor("DodgerBlue", 0.5, light); got != "DodgerBlue" {
		t.Errorf("unparseable colours pass through, got %s", got)
	}
}

func TestInkStrokeTapersAndStaysOnTheLine(t *testing.T) {
	line := []Pt{{0, 0}, {200, 0}}
	out := inkOutline(line, InkSpec{Width: 2, Seed: 7, TaperIn: 10, TaperOut: 10})
	if len(out) < 10 {
		t.Fatalf("expected a sampled outline, got %d points", len(out))
	}
	n := len(out) / 2
	widthAt := func(i int) float64 { return math.Abs(out[i].Y - out[len(out)-1-i].Y) }
	if widthAt(0) >= widthAt(n/2) || widthAt(n-1) >= widthAt(n/2) {
		t.Errorf("expected the stroke to be thinner at its ends: start %.2f, middle %.2f, end %.2f", widthAt(0), widthAt(n/2), widthAt(n-1))
	}
	for _, p := range out {
		if math.Abs(p.Y) > 3 {
			t.Fatalf("stroke wanders %.2fpx off a straight line", p.Y)
		}
	}
}

func TestThemeFromPropsAppliesLookAndSeed(t *testing.T) {
	th := ThemeFromProps(map[string]any{ThemeKey: "dark", LookKey: "watercolor", SeedKey: "db"})
	if !th.IsWatercolor() || th.Name != Dark.Name {
		t.Fatalf("expected the dark painted theme, got %+v", th)
	}
	if th.Seed != SeedFor("db") {
		t.Fatal("expected the seed to come from the node id")
	}
	if ThemeFromProps(map[string]any{}).IsWatercolor() {
		t.Fatal("expected the flat look by default")
	}

	dst := map[string]any{}
	CopyThemeProps(dst, map[string]any{ThemeKey: "light", LookKey: "watercolor", SeedKey: "db"}, "-read")
	if dst[LookKey] != "watercolor" || dst[SeedKey] != "db-read" || dst[ThemeKey] != "light" {
		t.Fatalf("unexpected copied props %v", dst)
	}
}
