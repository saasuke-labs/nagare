package core

import "math"

const FloatEpsilon = 0.0001

func NearlyEqual(a, b float64) bool {
	return math.Abs(a-b) <= FloatEpsilon
}

// GetX resolves an X coordinate from a percentage within a container.
func GetX(container Shape, percent float64) float64 {
	return container.X + (container.Width * clampPercent(percent))
}

// GetY resolves a Y coordinate from a percentage within a container.
func GetY(container Shape, percent float64) float64 {
	return container.Y + (container.Height * clampPercent(percent))
}

// GetWidth resolves width from a percentage of the container width.
func GetWidth(container Shape, percent float64) float64 {
	return container.Width * clampPercent(percent)
}

// GetHeight resolves height from a percentage of the container height.
func GetHeight(container Shape, percent float64) float64 {
	return container.Height * clampPercent(percent)
}

// InnerBox returns a child box relative to its parent box using percentage values.
func InnerBox(container Shape, xPercent, yPercent, widthPercent, heightPercent float64) Shape {
	return Shape{
		X:      GetX(container, xPercent),
		Y:      GetY(container, yPercent),
		Width:  GetWidth(container, widthPercent),
		Height: GetHeight(container, heightPercent),
	}
}

func clampPercent(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
