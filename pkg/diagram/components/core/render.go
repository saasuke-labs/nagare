package core

import (
	"strconv"
	"strings"

	"github.com/saasuke-labs/nagare/pkg/components"
)

func FloatProp(props map[string]any, key string, fallback float64) float64 {
	if props == nil {
		return fallback
	}
	v, ok := props[key]
	if !ok {
		return fallback
	}
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case string:
		return parseFloatOrDefault(t, fallback)
	default:
		return fallback
	}
}

func ShapeFromProps(props map[string]any, defaultWidth, defaultHeight float64) components.Shape {
	return components.Shape{
		X:      FloatProp(props, "x", 0),
		Y:      FloatProp(props, "y", 0),
		Width:  FloatProp(props, "w", defaultWidth),
		Height: FloatProp(props, "h", defaultHeight),
	}
}

func parseFloatOrDefault(input string, fallback float64) float64 {
	value, err := strconv.ParseFloat(strings.TrimSpace(input), 64)
	if err != nil {
		return fallback
	}
	return value
}
