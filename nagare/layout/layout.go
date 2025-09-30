package layout

import (
	"fmt"
	"math"
	"nagare/components"
	"nagare/parser"
	"nagare/props"
	"strings"
)

const (
	defaultBrowserWidth  = 640.0
	defaultBrowserHeight = 420.0
	defaultVMWidth       = 640.0
	defaultVMHeight      = 420.0
	defaultServerWidth   = 200.0
	defaultServerHeight  = 140.0
)

// Rect represents a rectangle in the layout
type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// Layout represents the computed layout of an element
type Layout struct {
	Bounds      Rect
	Children    []components.Component
	NodeIndex   map[string]components.Shape
	Connections []Arrow
}

// Point represents a 2D coordinate in canvas space.
type Point struct {
	X float64
	Y float64
}

// Arrow contains the resolved geometry for a parsed connection.
type Arrow struct {
	FromID      string
	ToID        string
	FromAnchor  string
	ToAnchor    string
	Start       Point
	End         Point
	BendPoints  []Point
	Style       string
	MarkerStart bool
	MarkerEnd   bool
}

type geometryProps struct {
	X      *int `prop:"x"`
	Y      *int `prop:"y"`
	Width  *int `prop:"w"`
	Height *int `prop:"h"`
}

func parseGeometry(def string) (geometryProps, error) {
	geom := geometryProps{}
	if strings.TrimSpace(def) == "" {
		return geom, nil
	}
	if err := props.ParseProps(def, &geom); err != nil {
		return geom, err
	}
	return geom, nil
}

func applyGeometry(shape *components.Shape, geom geometryProps) {
	if geom.Width != nil {
		shape.Width = float64(*geom.Width)
	}
	if geom.Height != nil {
		shape.Height = float64(*geom.Height)
	}
	if geom.X != nil {
		shape.X = float64(*geom.X)
	}
	if geom.Y != nil {
		shape.Y = float64(*geom.Y)
	}
}

// Calculate computes the layout for an AST
func Calculate(node parser.Node, canvasWidth, canvasHeight float64) Layout {
	boundsWidth := canvasWidth
	boundsHeight := canvasHeight
	nodeIndex := make(map[string]components.Shape)

	if layoutState, ok := node.Globals["layout"]; ok {
		if geom, err := parseGeometry(layoutState.PropsDef); err == nil {
			if geom.Width != nil {
				boundsWidth = float64(*geom.Width)
			}
			if geom.Height != nil {
				boundsHeight = float64(*geom.Height)
			}
		} else {
			fmt.Printf("failed to parse @layout props: %v\n", err)
		}
	}

	children := make([]components.Component, 0, len(node.Children))

	for _, child := range node.Children {
		switch child.Type {
		case "Browser":
			browser := components.NewBrowser()
			browser.Shape = components.Shape{
				Width:  defaultBrowserWidth,
				Height: defaultBrowserHeight,
				X:      0,
				Y:      0,
			}

			if idState, ok := child.States[child.Text]; ok {
				if geom, err := parseGeometry(idState.PropsDef); err == nil {
					applyGeometry(&browser.Shape, geom)
				} else {
					fmt.Printf("failed to parse geometry for %s: %v\n", child.Text, err)
				}

				if err := browser.Props.Parse(idState.PropsDef); err != nil {
					fmt.Printf("failed to parse props for %s: %v\n", child.Text, err)
				}
			}

			if child.State != "" {
				if state, ok := child.States[child.State]; ok {
					browser.State = state.Name
					if err := browser.Props.Parse(state.PropsDef); err != nil {
						fmt.Printf("failed to parse props for state %s: %v\n", state.Name, err)
					}
				}
			}

			children = append(children, browser)
			nodeIndex[child.Text] = browser.Shape
			fmt.Printf("State: %s, Props: %+v\n", browser.State, browser.Props)
		case "VM":
			vm := components.NewVM()
			vm.Shape = components.Shape{
				Width:  defaultVMWidth,
				Height: defaultVMHeight,
				X:      0,
				Y:      0,
			}

			if idState, ok := child.States[child.Text]; ok {
				if geom, err := parseGeometry(idState.PropsDef); err == nil {
					applyGeometry(&vm.Shape, geom)
				} else {
					fmt.Printf("failed to parse geometry for %s: %v\n", child.Text, err)
				}

				if err := vm.Props.Parse(idState.PropsDef); err != nil {
					fmt.Printf("failed to parse props for %s: %v\n", child.Text, err)
				}
			}

			if child.State != "" {
				if state, ok := child.States[child.State]; ok {
					vm.State = state.Name
					if err := vm.Props.Parse(state.PropsDef); err != nil {
						fmt.Printf("failed to parse props for state %s: %v\n", state.Name, err)
					}
				}
			}

			if len(child.Children) > 0 {
				childComponents := make([]components.Component, 0, len(child.Children))

				for _, grandchild := range child.Children {
					switch grandchild.Type {
					case "Server":
						server := components.NewServer(grandchild.Text)
						server.Shape = components.Shape{
							Width:  defaultServerWidth,
							Height: defaultServerHeight,
							X:      0,
							Y:      0,
						}

						if idState, ok := grandchild.States[grandchild.Text]; ok {
							if geom, err := parseGeometry(idState.PropsDef); err == nil {
								applyGeometry(&server.Shape, geom)
							} else {
								fmt.Printf("failed to parse geometry for %s: %v\n", grandchild.Text, err)
							}

							if err := server.Props.Parse(idState.PropsDef); err != nil {
								fmt.Printf("failed to parse props for %s: %v\n", grandchild.Text, err)
							}
						}

						if grandchild.State != "" {
							if state, ok := grandchild.States[grandchild.State]; ok {
								server.State = state.Name
								if geom, err := parseGeometry(state.PropsDef); err == nil {
									applyGeometry(&server.Shape, geom)
								} else {
									fmt.Printf("failed to parse geometry for state %s: %v\n", state.Name, err)
								}
								if err := server.Props.Parse(state.PropsDef); err != nil {
									fmt.Printf("failed to parse props for state %s: %v\n", state.Name, err)
								}
							}
						}

						childComponents = append(childComponents, server)

						absServerShape := server.Shape
						contentOffsetX := vm.Shape.Width * components.VMContentAreaXRatio
						contentOffsetY := vm.Shape.Height * components.VMContentAreaYRatio
						absServerShape.X = vm.Shape.X + contentOffsetX + absServerShape.X
						absServerShape.Y = vm.Shape.Y + contentOffsetY + absServerShape.Y
						nodeIndex[grandchild.Text] = absServerShape
					default:
						fmt.Printf("Unknown child type: %s\n", grandchild.Type)
					}
				}
				vm.Children = childComponents
			}

			children = append(children, vm)
			nodeIndex[child.Text] = vm.Shape
			fmt.Printf("State: %s, Props: %+v\n", vm.State, vm.Props)
		default:
			rect := &components.Rectangle{
				Shape: components.Shape{
					Width:  defaultServerWidth,
					Height: defaultServerHeight,
					X:      0,
					Y:      0,
				},
				Text: child.Text,
			}
			children = append(children, rect)
			nodeIndex[child.Text] = rect.Shape
		}
	}

	arrows := resolveConnections(node.Connections, nodeIndex)

	if len(arrows) > 0 {
		arrowComponents := make([]components.Component, 0, len(arrows))
		for _, arrow := range arrows {
			points := make([]components.Point, 0, len(arrow.BendPoints)+2)
			points = append(points, components.Point{X: arrow.Start.X, Y: arrow.Start.Y})
			for _, bend := range arrow.BendPoints {
				points = append(points, components.Point{X: bend.X, Y: bend.Y})
			}
			points = append(points, components.Point{X: arrow.End.X, Y: arrow.End.Y})

			arrowComponent := components.NewArrow(points)
			arrowComponent.Style = arrow.Style
			arrowComponent.MarkerStart = arrow.MarkerStart
			arrowComponent.MarkerEnd = arrow.MarkerEnd
			arrowComponents = append(arrowComponents, arrowComponent)
		}

		children = append(children, arrowComponents...)
	}

	return Layout{
		Bounds: Rect{
			X:      0,
			Y:      0,
			Width:  boundsWidth,
			Height: boundsHeight,
		},
		Children:    children,
		NodeIndex:   nodeIndex,
		Connections: arrows,
	}
}

func resolveConnections(connections []parser.Connection, nodeIndex map[string]components.Shape) []Arrow {
	arrows := make([]Arrow, 0, len(connections))
	for _, conn := range connections {
		fromShape, okFrom := nodeIndex[conn.FromID]
		toShape, okTo := nodeIndex[conn.ToID]
		if !okFrom || !okTo {
			fmt.Printf("connection skipped: missing endpoint %s -> %s\n", conn.FromID, conn.ToID)
			continue
		}

		fromAnchor := normalizeAnchor(conn.FromAnchor)
		toAnchor := normalizeAnchor(conn.ToAnchor)
		start := computeAnchorPoint(fromShape, fromAnchor)
		end := computeAnchorPoint(toShape, toAnchor)

		points := buildArrowPoints(start, end, fromAnchor, toAnchor)
		bendPoints := make([]Point, 0)
		if len(points) > 2 {
			bendPoints = append(bendPoints, points[1:len(points)-1]...)
		}

		arrows = append(arrows, Arrow{
			FromID:      conn.FromID,
			ToID:        conn.ToID,
			FromAnchor:  fromAnchor.Raw,
			ToAnchor:    toAnchor.Raw,
			Start:       points[0],
			End:         points[len(points)-1],
			BendPoints:  bendPoints,
			Style:       conn.Style,
			MarkerStart: false,
			MarkerEnd:   true,
		})
	}
	return arrows
}

func normalizeAnchor(anchor parser.AnchorDescriptor) parser.AnchorDescriptor {
	if anchor.Horizontal != 0 || anchor.Vertical != 0 || anchor.Raw == "" {
		return anchor
	}

	normalized := parser.AnchorDescriptor{Raw: anchor.Raw}
	lower := strings.ToLower(anchor.Raw)
	for _, r := range lower {
		switch r {
		case 'w':
			normalized.Horizontal = -1.0
		case 'e':
			normalized.Horizontal = 1.0
		case 'n':
			normalized.Vertical = -1.0
		case 's':
			normalized.Vertical = 1.0
		}
	}
	return normalized
}

func computeAnchorPoint(shape components.Shape, anchor parser.AnchorDescriptor) Point {
	point := Point{
		X: shape.X + shape.Width*0.5,
		Y: shape.Y + shape.Height*0.5,
	}

	switch {
	case anchor.Horizontal < 0:
		point.X = shape.X
		switch {
		case anchor.Vertical < 0:
			point.Y = shape.Y + shape.Height*0.25
		case anchor.Vertical > 0:
			point.Y = shape.Y + shape.Height*0.75
		default:
			point.Y = shape.Y + shape.Height*0.5
		}
	case anchor.Horizontal > 0:
		point.X = shape.X + shape.Width
		switch {
		case anchor.Vertical < 0:
			point.Y = shape.Y + shape.Height*0.25
		case anchor.Vertical > 0:
			point.Y = shape.Y + shape.Height*0.75
		default:
			point.Y = shape.Y + shape.Height*0.5
		}
	default:
		switch {
		case anchor.Vertical < 0:
			point.Y = shape.Y
			point.X = shape.X + shape.Width*0.5
		case anchor.Vertical > 0:
			point.Y = shape.Y + shape.Height
			point.X = shape.X + shape.Width*0.5
		default:
			point.X = shape.X + shape.Width*0.5
			point.Y = shape.Y + shape.Height*0.5
		}
	}

	return point
}

func buildArrowPoints(start, end Point, fromAnchor, toAnchor parser.AnchorDescriptor) []Point {
	points := []Point{start}

	if almostEqual(start.X, end.X) || almostEqual(start.Y, end.Y) {
		points = append(points, end)
		return points
	}

	const elbowPadding = 24.0

	horizontalFirst := determineFirstAxis(fromAnchor, toAnchor)

	if horizontalFirst {
		direction := axisDirection(fromAnchor.Horizontal, toAnchor.Horizontal)
		elbowX := start.X + direction*elbowPadding
		points = append(points, Point{X: elbowX, Y: start.Y})
		points = append(points, Point{X: elbowX, Y: end.Y})
	} else {
		direction := axisDirection(fromAnchor.Vertical, toAnchor.Vertical)
		elbowY := start.Y + direction*elbowPadding
		points = append(points, Point{X: start.X, Y: elbowY})
		points = append(points, Point{X: end.X, Y: elbowY})
	}

	points = append(points, end)
	return points
}

func determineFirstAxis(fromAnchor, toAnchor parser.AnchorDescriptor) bool {
	if fromAnchor.Horizontal != 0 {
		return true
	}
	if fromAnchor.Vertical != 0 {
		return false
	}
	if toAnchor.Horizontal != 0 {
		return true
	}
	return false
}

func axisDirection(primary, secondary float64) float64 {
	if primary < 0 {
		return -1
	}
	if primary > 0 {
		return 1
	}
	if secondary < 0 {
		return -1
	}
	if secondary > 0 {
		return 1
	}
	return 1
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.0001
}
