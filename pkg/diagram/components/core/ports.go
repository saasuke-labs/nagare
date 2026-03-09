package core

import "fmt"

const (
	PortNorth  = "n"
	PortSouth  = "s"
	PortEast   = "e"
	PortWest   = "w"
	PortCenter = "c"
)

// RectPort returns common rectangle ports for a bounding box.
func RectPort(box Shape, name string) (Point, error) {
	switch name {
	case PortNorth:
		return Point{X: box.X + box.Width/2, Y: box.Y}, nil
	case PortSouth:
		return Point{X: box.X + box.Width/2, Y: box.Y + box.Height}, nil
	case PortEast:
		return Point{X: box.X + box.Width, Y: box.Y + box.Height/2}, nil
	case PortWest:
		return Point{X: box.X, Y: box.Y + box.Height/2}, nil
	case PortCenter:
		return Point{X: box.X + box.Width/2, Y: box.Y + box.Height/2}, nil
	default:
		return Point{}, fmt.Errorf("unsupported port %q", name)
	}
}
