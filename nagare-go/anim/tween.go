package anim

import (
	"math"
	"nagare-go/model"
)

// ----- types -----

// ----- public sampling -----

// AtNumber returns value at time t from numeric keyframes (sorted by T).
func AtNumber(kfs []model.KfN, t float64) float64 {
	n := len(kfs)
	if n == 0 {
		return 0
	}
	if t <= kfs[0].T {
		return kfs[0].V
	}
	for i := 0; i < n-1; i++ {
		a, b := kfs[i], kfs[i+1]
		if t <= b.T {
			if a.Spring != nil {
				// spring from a.V -> b.V over segment duration
				dt := b.T - a.T
				u := clamp01((t - a.T) / dt)
				return springStep(a.V, b.V, a.Spring, a.VelocityOut, dt*u)
			}
			u := clamp01((t - a.T) / (b.T - a.T))
			w := applyEase(u, a.Ease, a.EaseBezier)
			return lerp(a.V, b.V, w)
		}
	}
	return kfs[n-1].V
}

// ----- easing curves -----

func applyEase(u float64, k model.EaseKind, bez *model.Bezier) float64 {
	if bez != nil {
		return cubicBezier(bez.X1, bez.Y1, bez.X2, bez.Y2, u)
	}
	switch k {
	case model.EaseIn:
		return u * u
	case model.EaseOut:
		return 1 - (1-u)*(1-u)
	case model.EaseInOut:
		if u < 0.5 {
			return 2 * u * u
		}
		return 1 - math.Pow(-2*u+2, 2)/2
	default:
		return u
	}
}

// Solve cubic-bezier x(t)=u for t, then return y(t).
func cubicBezier(x1, y1, x2, y2, u float64) float64 {
	t := u
	for i := 0; i < 5; i++ {
		x := bezCoord(x1, x2, t)
		dx := bezSlope(x1, x2, t)
		if dx == 0 {
			break
		}
		t -= (x - u) / dx
		t = clamp01(t)
	}
	return bezCoord(y1, y2, t)
}

func bezCoord(p1, p2, t float64) float64 {
	mt := 1 - t
	return 3*mt*mt*t*p1 + 3*mt*t*t*p2 + t*t*t
}
func bezSlope(p1, p2, t float64) float64 {
	mt := 1 - t
	return 3*mt*mt*p1 + 6*mt*t*(p2-p1) + 3*t*t*(1-p2)
}

// ----- spring (closed form DHO) -----

func springStep(a, b float64, s *model.Spring, vel0 float64, tau float64) float64 {
	if s == nil {
		return b
	}
	m := s.Mass
	if m <= 0 {
		m = 1
	}
	k := s.K
	if k <= 0 {
		k = 100
	}
	z := s.Zeta
	omega0 := math.Sqrt(k / m)

	x0 := a
	xf := b
	y0 := x0 - xf
	v0 := vel0

	switch {
	case z < 1: // underdamped
		wd := omega0 * math.Sqrt(1-z*z)
		A := y0
		B := (v0 + z*omega0*y0) / wd
		exp := math.Exp(-z * omega0 * tau)
		y := exp * (A*math.Cos(wd*tau) + B*math.Sin(wd*tau))
		return xf + y
	case z == 1: // critically damped
		exp := math.Exp(-omega0 * tau)
		y := (y0 + (v0+omega0*y0)*tau) * exp
		return xf + y
	default: // overdamped
		sq := math.Sqrt(z*z - 1)
		r1 := -omega0 * (z - sq)
		r2 := -omega0 * (z + sq)
		C1 := (v0 - r2*y0) / (r1 - r2)
		C2 := y0 - C1
		y := C1*math.Exp(r1*tau) + C2*math.Exp(r2*tau)
		return xf + y
	}
}

// ----- utils -----

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
func lerp(a, b, t float64) float64 { return a + (b-a)*t }
