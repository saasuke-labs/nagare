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

// SVGDrawer is compatible with the current component contract in pkg/components.
type SVGDrawer interface {
	Draw() string
}

// DrawerAdapter bridges legacy components into the new core.Component contract.
type DrawerAdapter struct {
	ID     string
	Box    BoundingBox
	Drawer SVGDrawer
	Ports  map[string]Point
	Parent BoundingBox
}

func (a *DrawerAdapter) Draw() (string, error) {
	if a == nil || a.Drawer == nil {
		return "", fmt.Errorf("drawer adapter has no component")
	}
	return a.Drawer.Draw(), nil
}

func (a *DrawerAdapter) BoundingBox() BoundingBox {
	if a == nil {
		return BoundingBox{}
	}
	return a.Box
}

func (a *DrawerAdapter) Port(name string) (Point, error) {
	if a == nil {
		return Point{}, fmt.Errorf("drawer adapter is nil")
	}
	if len(a.Ports) > 0 {
		if p, ok := a.Ports[name]; ok {
			return p, nil
		}
	}
	return Point{}, fmt.Errorf("port %q not found", name)
}

func (a *DrawerAdapter) SetContainer(box BoundingBox) {
	if a == nil {
		return
	}
	a.Parent = box
}

func (a *DrawerAdapter) SetBoundingBox(box BoundingBox) {
	if a == nil {
		return
	}
	a.Box = box
}

func (a *DrawerAdapter) ApplyProps(raw string) error {
	_ = raw
	return nil
}
