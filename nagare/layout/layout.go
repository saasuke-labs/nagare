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
	defaultComponentX    = 0.0
	defaultComponentY    = 0.0
	arrowElbowPadding    = 24.0
	floatEqualityEpsilon = 0.0001
)

const (
	componentTypeBrowser = "Browser"
	componentTypeVM      = "VM"
	componentTypeServer  = "Server"
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

type propertyParser interface {
	Parse(string) error
}

func parseGeometryProps(def string) (geometryProps, error) {
	geom := geometryProps{}
	if strings.TrimSpace(def) == "" {
		return geom, nil
	}
	if err := props.ParseProps(def, &geom); err != nil {
		return geom, err
	}
	return geom, nil
}

func applyGeometryProps(shape *components.Shape, geom geometryProps) {
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
	boundsWidth, boundsHeight := calculateCanvasBounds(node, canvasWidth, canvasHeight)
	nodeIndex := make(map[string]components.Shape)

	children := make([]components.Component, 0, len(node.Children))
	for _, child := range node.Children {
		children = append(children, buildComponentTree(child, nodeIndex)...)
	}

	arrows := resolveConnections(node.Connections, nodeIndex)
	if len(arrows) > 0 {
		children = append(children, buildArrowComponents(arrows)...)
	}

	return Layout{
		Bounds: Rect{
			X:      defaultComponentX,
			Y:      defaultComponentY,
			Width:  boundsWidth,
			Height: boundsHeight,
		},
		Children:    children,
		NodeIndex:   nodeIndex,
		Connections: arrows,
	}
}

func calculateCanvasBounds(node parser.Node, defaultWidth, defaultHeight float64) (float64, float64) {
	boundsWidth := defaultWidth
	boundsHeight := defaultHeight

	layoutState, ok := node.Globals["layout"]
	if !ok {
		return boundsWidth, boundsHeight
	}

	geometry, err := parseGeometryProps(layoutState.PropsDef)
	if err != nil {
		fmt.Printf("failed to parse @layout props: %v\n", err)
		return boundsWidth, boundsHeight
	}

	if geometry.Width != nil {
		boundsWidth = float64(*geometry.Width)
	}
	if geometry.Height != nil {
		boundsHeight = float64(*geometry.Height)
	}

	return boundsWidth, boundsHeight
}

func buildComponentTree(node parser.Node, nodeIndex map[string]components.Shape) []components.Component {
	switch node.Type {
	case componentTypeBrowser:
		return []components.Component{buildBrowser(node, nodeIndex)}
	case componentTypeVM:
		return []components.Component{buildVM(node, nodeIndex)}
	default:
		return []components.Component{buildFallbackRectangle(node, nodeIndex)}
	}
}

func buildBrowser(node parser.Node, nodeIndex map[string]components.Shape) components.Component {
	browser := components.NewBrowser()
	browser.Text = node.Text
	browser.Shape = components.Shape{
		Width:  defaultBrowserWidth,
		Height: defaultBrowserHeight,
		X:      defaultComponentX,
		Y:      defaultComponentY,
	}

	applyIDStateProperties(node, &browser.Shape, &browser.Props, node.Text)
	browser.State = applyNamedStateProperties(node, &browser.Shape, &browser.Props, false)

	nodeIndex[node.Text] = browser.Shape
	fmt.Printf("State: %s, Props: %+v\n", browser.State, browser.Props)
	return browser
}

func buildVM(node parser.Node, nodeIndex map[string]components.Shape) components.Component {
	vm := components.NewVM()
	vm.Text = node.Text
	vm.Shape = components.Shape{
		Width:  defaultVMWidth,
		Height: defaultVMHeight,
		X:      defaultComponentX,
		Y:      defaultComponentY,
	}

	applyIDStateProperties(node, &vm.Shape, &vm.Props, node.Text)
	vm.State = applyNamedStateProperties(node, &vm.Shape, &vm.Props, false)

	layoutVMChildren(node, vm, nodeIndex)
	nodeIndex[node.Text] = vm.Shape
	fmt.Printf("State: %s, Props: %+v\n", vm.State, vm.Props)
	return vm
}

func layoutVMChildren(parent parser.Node, vm *components.VM, nodeIndex map[string]components.Shape) {
	if len(parent.Children) == 0 {
		return
	}

	for _, child := range parent.Children {
		switch child.Type {
		case componentTypeServer:
			server := buildServer(child, vm, nodeIndex)
			vm.AddChild(server)
		default:
			fmt.Printf("Unknown child type: %s\n", child.Type)
		}
	}
}

func buildServer(node parser.Node, vm *components.VM, nodeIndex map[string]components.Shape) *components.Server {
	server := components.NewServer(node.Text)
	server.Shape = components.Shape{
		Width:  defaultServerWidth,
		Height: defaultServerHeight,
		X:      defaultComponentX,
		Y:      defaultComponentY,
	}

	applyIDStateProperties(node, &server.Shape, &server.Props, node.Text)
	server.State = applyNamedStateProperties(node, &server.Shape, &server.Props, true)

	absServerShape := server.Shape
	contentOffsetX := vm.Shape.Width * components.VMContentAreaXRatio
	contentOffsetY := vm.Shape.Height * components.VMContentAreaYRatio
	absServerShape.X = vm.Shape.X + contentOffsetX + absServerShape.X
	absServerShape.Y = vm.Shape.Y + contentOffsetY + absServerShape.Y
	nodeIndex[node.Text] = absServerShape

	return server
}

func buildFallbackRectangle(node parser.Node, nodeIndex map[string]components.Shape) components.Component {
	rect := &components.Rectangle{
		Shape: components.Shape{
			Width:  defaultServerWidth,
			Height: defaultServerHeight,
			X:      defaultComponentX,
			Y:      defaultComponentY,
		},
		Text: node.Text,
	}
	nodeIndex[node.Text] = rect.Shape
	return rect
}

func applyIDStateProperties(node parser.Node, shape *components.Shape, props propertyParser, componentID string) {
	idState, ok := node.States[node.Text]
	if !ok {
		return
	}

	applyGeometryDefinition(componentID, shape, idState.PropsDef)
	parseComponentProps(componentID, props, idState.PropsDef)
}

func applyNamedStateProperties(node parser.Node, shape *components.Shape, props propertyParser, includeGeometry bool) string {
	if node.State == "" {
		return ""
	}

	state, ok := node.States[node.State]
	if !ok {
		return ""
	}

	if includeGeometry {
		applyGeometryDefinition(fmt.Sprintf("state %s", state.Name), shape, state.PropsDef)
	}
	parseComponentProps(fmt.Sprintf("state %s", state.Name), props, state.PropsDef)
	return state.Name
}

func applyGeometryDefinition(target string, shape *components.Shape, propsDef string) {
	if shape == nil {
		return
	}

	geometry, err := parseGeometryProps(propsDef)
	if err != nil {
		fmt.Printf("failed to parse geometry for %s: %v\n", target, err)
		return
	}
	applyGeometryProps(shape, geometry)
}

func parseComponentProps(target string, parser propertyParser, propsDef string) {
	if parser == nil {
		return
	}
	if err := parser.Parse(propsDef); err != nil {
		fmt.Printf("failed to parse props for %s: %v\n", target, err)
	}
}

func buildArrowComponents(arrows []Arrow) []components.Component {
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
	return arrowComponents
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

		points := routeArrowPoints(start, end, fromAnchor, toAnchor)
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

func routeArrowPoints(start, end Point, fromAnchor, toAnchor parser.AnchorDescriptor) []Point {
	points := []Point{start}

	if floatsNearlyEqual(start.X, end.X) || floatsNearlyEqual(start.Y, end.Y) {
		points = append(points, end)
		return points
	}

	horizontalFirst := shouldRouteHorizontallyFirst(fromAnchor, toAnchor)

	if horizontalFirst {
		direction := resolveAxisDirection(fromAnchor.Horizontal, toAnchor.Horizontal)
		elbowX := start.X + direction*arrowElbowPadding
		points = append(points, Point{X: elbowX, Y: start.Y})
		points = append(points, Point{X: elbowX, Y: end.Y})
	} else {
		direction := resolveAxisDirection(fromAnchor.Vertical, toAnchor.Vertical)
		elbowY := start.Y + direction*arrowElbowPadding
		points = append(points, Point{X: start.X, Y: elbowY})
		points = append(points, Point{X: end.X, Y: elbowY})
	}

	points = append(points, end)
	return points
}

func shouldRouteHorizontallyFirst(fromAnchor, toAnchor parser.AnchorDescriptor) bool {
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

func resolveAxisDirection(primary, secondary float64) float64 {
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

func floatsNearlyEqual(a, b float64) bool {
	return math.Abs(a-b) < floatEqualityEpsilon
}
