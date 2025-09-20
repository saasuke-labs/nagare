package model

type EaseKind int

const (
	Linear EaseKind = iota
	EaseIn
	EaseOut
	EaseInOut
)

// Cubic-bezier control points (CSS-like)
type Bezier struct{ X1, Y1, X2, Y2 float64 }

// Simple spring (damped harmonic oscillator params)
type Spring struct {
	K    float64 // stiffness (suggest 60..200)
	Zeta float64 // damping ratio (0.5..1.2 nice)
	Mass float64 // mass (>= 0, usually 1)
}

type KfN struct {
	T           float64 `json:"T"`
	V           float64 `json:"T"`
	Ease        EaseKind
	EaseBezier  *Bezier
	Spring      *Spring
	VelocityOut float64
}

type Node struct {
	ID      string  `json:"id"`
	Label   string  `json:"label,omitempty"`
	Type    string  `json:"type,omitempty"`
	Parent  string  `json:"parent,omitempty"`
	X       float64 `json:"x,omitempty"`
	Y       float64 `json:"y,omitempty"`
	W       float64 `json:"w,omitempty"`
	H       float64 `json:"h,omitempty"`
	XTrack  []KfN   `json:"xTrack,omitempty"`
	YTrack  []KfN   `json:"yTrack,omitempty"`
	Opacity []KfN   `json:"opacity,omitempty"`
}

type Edge struct {
	ID     string `json:"id,omitempty"`
	From   string `json:"from"`
	To     string `json:"to"`
	Label  string `json:"label,omitempty"`
	FlowOn []KfN  `json:"flowOn,omitempty"`
}

type Scene struct {
	Width       int         `json:"width,omitempty"`
	Height      int         `json:"height,omitempty"`
	FPS         float64     `json:"fps,omitempty"`
	DurationSec float64     `json:"durationSec"`
	Nodes       []SceneNode `json:"nodes"`
	Edges       []SceneEdge `json:"edges"`
	Bg          string      `json:"bg,omitempty"`
	Fg          string      `json:"fg,omitempty"`
}

type SceneNode struct {
	ID    string  `json:"id"`
	Label string  `json:"label,omitempty"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	W     float64 `json:"w"`
	H     float64 `json:"h"`
	// simple tracks (optional)
	XTrack  []KfN `json:"xTrack,omitempty"`
	YTrack  []KfN `json:"yTrack,omitempty"`
	Opacity []KfN `json:"opacity,omitempty"`
}

type SceneEdge struct {
	ID     string `json:"id"`
	From   string `json:"from"`
	To     string `json:"to"`
	Label  string `json:"label,omitempty"`
	FlowOn []KfN  `json:"flowOn,omitempty"`
}
