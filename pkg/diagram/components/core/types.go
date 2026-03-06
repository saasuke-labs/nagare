package core

import "fmt"

// Point represents a coordinate in diagram canvas space.
type Point struct {
	X float64
	Y float64
}

// BoundingBox represents an absolute rectangle in diagram canvas space.
type BoundingBox struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// Component defines the target interface for diagram components.
// Implementations own rendering, geometry, and port resolution behavior.
type Component interface {
	Draw() (string, error)
	BoundingBox() BoundingBox
	Port(name string) (Point, error)
}

// ConfigurableComponent extends Component with initialization hooks used by layout orchestration.
type ConfigurableComponent interface {
	Component
	SetContainer(BoundingBox)
	SetBoundingBox(BoundingBox)
	ApplyProps(raw string) error
}

// LegacyDrawer is compatible with the current component contract in pkg/components.
type LegacyDrawer interface {
	Draw() string
}

// LegacyAdapter bridges legacy components into the new core.Component contract.
type LegacyAdapter struct {
	ID     string
	Box    BoundingBox
	Legacy LegacyDrawer
	Ports  map[string]Point
	Parent BoundingBox
}

func (a *LegacyAdapter) Draw() (string, error) {
	if a == nil || a.Legacy == nil {
		return "", fmt.Errorf("legacy adapter has no component")
	}
	return a.Legacy.Draw(), nil
}

func (a *LegacyAdapter) BoundingBox() BoundingBox {
	if a == nil {
		return BoundingBox{}
	}
	return a.Box
}

func (a *LegacyAdapter) Port(name string) (Point, error) {
	if a == nil {
		return Point{}, fmt.Errorf("legacy adapter is nil")
	}
	if len(a.Ports) > 0 {
		if p, ok := a.Ports[name]; ok {
			return p, nil
		}
	}
	return Point{}, fmt.Errorf("port %q not found", name)
}

func (a *LegacyAdapter) SetContainer(box BoundingBox) {
	if a == nil {
		return
	}
	a.Parent = box
}

func (a *LegacyAdapter) SetBoundingBox(box BoundingBox) {
	if a == nil {
		return
	}
	a.Box = box
}

func (a *LegacyAdapter) ApplyProps(raw string) error {
	_ = raw
	return nil
}
