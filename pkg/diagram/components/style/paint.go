package style

// Seeded geometry for the watercolor look: a small PRNG, 1D value noise,
// outline builders and the polygon deformation behind layered washes.
// Everything here is pure and deterministic: the same seed gives the same
// SVG, so diagrams diff cleanly and golden tests stay stable.

import (
	"hash/fnv"
	"math"
)

// Pt is a point in a component's local space.
type Pt struct{ X, Y float64 }

// SeedFor derives a stable seed from any number of strings (usually a node id
// plus a role such as "wash" or "ink").
func SeedFor(parts ...string) uint64 {
	h := fnv.New64a()
	for _, p := range parts {
		_, _ = h.Write([]byte(p))
		_, _ = h.Write([]byte{0})
	}
	return h.Sum64()
}

// rng is splitmix64: tiny, fast and stable across Go versions.
type rng struct{ s uint64 }

func newRng(seed uint64) *rng { return &rng{s: seed ^ 0x9e3779b97f4a7c15} }

func (r *rng) next() uint64 {
	r.s += 0x9e3779b97f4a7c15
	z := r.s
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// float returns a value in [0, 1).
func (r *rng) float() float64 { return float64(r.next()>>11) / float64(1<<53) }

// signed returns a value in [-1, 1).
func (r *rng) signed() float64 { return r.float()*2 - 1 }

// gauss returns a standard normal sample, clipped to ±3 so a rare outlier
// never throws an edge across the page.
func (r *rng) gauss() float64 {
	u := math.Max(1e-9, r.float())
	v := r.float()
	g := math.Sqrt(-2*math.Log(u)) * math.Cos(2*math.Pi*v)
	return Clamp(g, -3, 3)
}

func hash01(seed uint64, i int64) float64 {
	r := rng{s: seed ^ uint64(i)*0xd1b54a32d192ed03}
	return r.float()
}

// noise1 is smooth 1D value noise in [-1, 1].
func noise1(seed uint64, x float64) float64 {
	i := math.Floor(x)
	f := x - i
	u := f * f * (3 - 2*f)
	a := hash01(seed, int64(i))
	b := hash01(seed, int64(i)+1)
	return (a+(b-a)*u)*2 - 1
}

// fbm1 sums three octaves of noise1.
func fbm1(seed uint64, x float64) float64 {
	return (noise1(seed, x) + 0.5*noise1(seed+1, x*2.03) + 0.25*noise1(seed+2, x*4.01)) / 1.75
}

// RoundedRect returns a closed outline of a rounded rectangle, sampled so
// that every edge has points along it (deformation needs them).
func RoundedRect(x, y, w, h, r float64, step float64) []Pt {
	r = Clamp(r, 0, math.Min(w, h)/2)
	if step <= 0 {
		step = 24
	}
	var pts []Pt
	edge := func(ax, ay, bx, by float64) {
		n := int(math.Max(1, math.Round(math.Hypot(bx-ax, by-ay)/step)))
		for i := 0; i < n; i++ {
			t := float64(i) / float64(n)
			pts = append(pts, Pt{ax + (bx-ax)*t, ay + (by-ay)*t})
		}
	}
	arc := func(cx, cy, a0 float64) {
		if r <= 0 {
			return
		}
		const n = 3
		for i := 0; i < n; i++ {
			a := a0 + float64(i)/n*math.Pi/2
			pts = append(pts, Pt{cx + math.Cos(a)*r, cy + math.Sin(a)*r})
		}
	}
	edge(x+r, y, x+w-r, y)
	arc(x+w-r, y+r, -math.Pi/2)
	edge(x+w, y+r, x+w, y+h-r)
	arc(x+w-r, y+h-r, 0)
	edge(x+w-r, y+h, x+r, y+h)
	arc(x+r, y+h-r, math.Pi/2)
	edge(x, y+h-r, x, y+r)
	arc(x+r, y+r, math.Pi)
	return pts
}

// Ellipse returns a closed elliptical outline.
func Ellipse(cx, cy, rx, ry float64, n int) []Pt {
	if n < 8 {
		n = 8
	}
	pts := make([]Pt, n)
	for i := range pts {
		a := float64(i) / float64(n) * 2 * math.Pi
		pts[i] = Pt{cx + math.Cos(a)*rx, cy + math.Sin(a)*ry}
	}
	return pts
}

// deform runs recursive midpoint displacement on a closed polygon: the
// classic way to build a watercolor wash from many slightly different layers.
// variance is relative to each edge's length, so long edges wander more.
func deform(poly []Pt, depth int, variance, maxOff float64, r *rng) []Pt {
	cur := poly
	for level := 0; level < depth; level++ {
		next := make([]Pt, 0, len(cur)*2)
		for i, a := range cur {
			b := cur[(i+1)%len(cur)]
			dx, dy := b.X-a.X, b.Y-a.Y
			l := math.Hypot(dx, dy)
			if l < 1e-6 {
				next = append(next, a)
				continue
			}
			off := Clamp(r.gauss()*variance*l*0.5, -maxOff, maxOff)
			nx, ny := -dy/l, dx/l
			next = append(next, a, Pt{(a.X+b.X)/2 + nx*off, (a.Y+b.Y)/2 + ny*off})
		}
		cur = next
	}
	return cur
}

// jitter moves every vertex a little, independently.
func jitter(poly []Pt, amount float64, r *rng) []Pt {
	out := make([]Pt, len(poly))
	for i, p := range poly {
		out[i] = Pt{p.X + r.gauss()*amount, p.Y + r.gauss()*amount}
	}
	return out
}

// resample walks a polyline and returns points every step px of arc length.
// closed adds the closing segment.
func resample(pts []Pt, step float64, closed bool) []Pt {
	if len(pts) < 2 {
		return pts
	}
	src := pts
	if closed {
		src = append(append([]Pt{}, pts...), pts[0])
	}
	out := []Pt{src[0]}
	carry := 0.0
	for i := 1; i < len(src); i++ {
		a, b := src[i-1], src[i]
		l := math.Hypot(b.X-a.X, b.Y-a.Y)
		if l < 1e-9 {
			continue
		}
		d := step - carry
		for d <= l {
			t := d / l
			out = append(out, Pt{a.X + (b.X-a.X)*t, a.Y + (b.Y-a.Y)*t})
			d += step
		}
		carry = l - (d - step)
	}
	last := src[len(src)-1]
	if p := out[len(out)-1]; math.Hypot(p.X-last.X, p.Y-last.Y) > step*0.25 {
		out = append(out, last)
	}
	if closed && len(out) > 1 {
		// The walk ends where it started; drop the duplicate.
		if p := out[len(out)-1]; math.Hypot(p.X-out[0].X, p.Y-out[0].Y) < step*0.5 {
			out = out[:len(out)-1]
		}
	}
	return out
}

func polyLength(pts []Pt) float64 {
	l := 0.0
	for i := 1; i < len(pts); i++ {
		l += math.Hypot(pts[i].X-pts[i-1].X, pts[i].Y-pts[i-1].Y)
	}
	return l
}

// roundCorners replaces each interior vertex of a polyline with a short
// quadratic bend, so orthogonal routes read as drawn, not plotted.
func roundCorners(pts []Pt, radius float64) []Pt {
	if len(pts) < 3 {
		return pts
	}
	out := []Pt{pts[0]}
	for i := 1; i < len(pts)-1; i++ {
		a, p, b := pts[i-1], pts[i], pts[i+1]
		la := math.Hypot(p.X-a.X, p.Y-a.Y)
		lb := math.Hypot(b.X-p.X, b.Y-p.Y)
		r := math.Min(radius, math.Min(la, lb)*0.45)
		if r < 1 {
			out = append(out, p)
			continue
		}
		s := Pt{p.X + (a.X-p.X)/la*r, p.Y + (a.Y-p.Y)/la*r}
		e := Pt{p.X + (b.X-p.X)/lb*r, p.Y + (b.Y-p.Y)/lb*r}
		const n = 6
		for k := 0; k <= n; k++ {
			t := float64(k) / n
			mt := 1 - t
			out = append(out, Pt{
				mt*mt*s.X + 2*mt*t*p.X + t*t*e.X,
				mt*mt*s.Y + 2*mt*t*p.Y + t*t*e.Y,
			})
		}
	}
	return append(out, pts[len(pts)-1])
}
