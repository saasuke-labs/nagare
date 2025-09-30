package layout

import (
	"testing"

	"nagare/components"
	"nagare/parser"
)

type stubProps struct {
	last string
}

func (s *stubProps) Parse(input string) error {
	s.last = input
	return nil
}

func TestCalculateStoresAbsoluteShapesAndConnections(t *testing.T) {
	root := parser.Node{
		Globals: make(map[string]parser.State),
		Children: []parser.Node{
			{
				Type: "VM",
				Text: "vm1",
				States: map[string]parser.State{
					"vm1": {Name: "vm1", PropsDef: "x:100,y:50,w:640,h:420"},
				},
				Children: []parser.Node{
					{
						Type: "Server",
						Text: "srv1",
						States: map[string]parser.State{
							"srv1": {Name: "srv1", PropsDef: "x:30,y:40"},
						},
					},
				},
			},
		},
		Connections: []parser.Connection{
			{
				FromID: "srv1",
				FromAnchor: parser.AnchorDescriptor{
					Raw:        "e",
					Horizontal: 1,
				},
				ToID: "vm1",
				ToAnchor: parser.AnchorDescriptor{
					Raw:        "w",
					Horizontal: -1,
				},
			},
			{
				FromID: "srv1",
				FromAnchor: parser.AnchorDescriptor{
					Raw:        "wn",
					Horizontal: -1,
					Vertical:   -1,
				},
				ToID: "vm1",
				ToAnchor: parser.AnchorDescriptor{
					Raw:        "en",
					Horizontal: 1,
					Vertical:   -1,
				},
			},
		},
	}

	result := Calculate(root, 1024, 768)

	if len(result.Children) != len(root.Children)+len(root.Connections) {
		expected := len(root.Children) + len(root.Connections)
		t.Fatalf("expected %d components, got %d", expected, len(result.Children))
	}

	arrowStart := len(result.Children) - len(root.Connections)
	if arrowStart < 0 {
		t.Fatalf("unexpected arrow start index %d", arrowStart)
	}

	for i := 0; i < arrowStart; i++ {
		if _, ok := result.Children[i].(*components.Arrow); ok {
			t.Fatalf("expected child %d to be a non-arrow component", i)
		}
	}
	for i := arrowStart; i < len(result.Children); i++ {
		if _, ok := result.Children[i].(*components.Arrow); !ok {
			t.Fatalf("expected child %d to be an arrow component", i)
		}
	}

	serverShape, ok := result.NodeIndex["srv1"]
	if !ok {
		t.Fatalf("expected srv1 shape in node index")
	}

	vmShape, ok := result.NodeIndex["vm1"]
	if !ok {
		t.Fatalf("expected vm1 shape in node index")
	}

	contentOffsetX := vmShape.Width * components.VMContentAreaXRatio
	contentOffsetY := vmShape.Height * components.VMContentAreaYRatio

	expectedServerX := vmShape.X + contentOffsetX + 30
	expectedServerY := vmShape.Y + contentOffsetY + 40

	if serverShape.X != expectedServerX {
		t.Fatalf("expected server X %f, got %f", expectedServerX, serverShape.X)
	}
	if serverShape.Y != expectedServerY {
		t.Fatalf("expected server Y %f, got %f", expectedServerY, serverShape.Y)
	}

	if len(result.Connections) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(result.Connections))
	}

	first := result.Connections[0]
	if first.Start.X != expectedServerX+serverShape.Width {
		t.Fatalf("expected first connection start X %f, got %f", expectedServerX+serverShape.Width, first.Start.X)
	}
	if first.Start.Y != expectedServerY+serverShape.Height/2 {
		t.Fatalf("expected first connection start Y %f, got %f", expectedServerY+serverShape.Height/2, first.Start.Y)
	}
	if first.End.X != vmShape.X {
		t.Fatalf("expected first connection end X %f, got %f", vmShape.X, first.End.X)
	}
	if first.End.Y != vmShape.Y+vmShape.Height/2 {
		t.Fatalf("expected first connection end Y %f, got %f", vmShape.Y+vmShape.Height/2, first.End.Y)
	}

	second := result.Connections[1]
	if second.Start.X != expectedServerX {
		t.Fatalf("expected second connection start X %f, got %f", expectedServerX, second.Start.X)
	}
	if second.Start.Y != expectedServerY+serverShape.Height*0.25 {
		t.Fatalf("expected second connection start Y %f, got %f", expectedServerY+serverShape.Height*0.25, second.Start.Y)
	}
	if second.End.X != vmShape.X+vmShape.Width {
		t.Fatalf("expected second connection end X %f, got %f", vmShape.X+vmShape.Width, second.End.X)
	}
	if second.End.Y != vmShape.Y+vmShape.Height*0.25 {
		t.Fatalf("expected second connection end Y %f, got %f", vmShape.Y+vmShape.Height*0.25, second.End.Y)
	}
}

func TestCalculateUsesLayoutGlobalOverrides(t *testing.T) {
	root := parser.Node{
		Globals: map[string]parser.State{
			"layout": {Name: "layout", PropsDef: "w:800,h:600"},
		},
	}

	result := Calculate(root, 1024, 768)

	if result.Bounds.Width != 800 {
		t.Fatalf("expected bounds width 800, got %f", result.Bounds.Width)
	}
	if result.Bounds.Height != 600 {
		t.Fatalf("expected bounds height 600, got %f", result.Bounds.Height)
	}
}

func TestRouteArrowPointsRespectsAnchorPriority(t *testing.T) {
	start := Point{X: 10, Y: 10}
	end := Point{X: 110, Y: 80}
	fromAnchor := parser.AnchorDescriptor{Horizontal: 1}
	toAnchor := parser.AnchorDescriptor{Vertical: 1}

	points := routeArrowPoints(start, end, fromAnchor, toAnchor)

	expected := []Point{
		start,
		{X: start.X + arrowElbowPadding, Y: start.Y},
		{X: start.X + arrowElbowPadding, Y: end.Y},
		end,
	}

	if len(points) != len(expected) {
		t.Fatalf("expected %d points, got %d", len(expected), len(points))
	}

	for i := range points {
		if !floatsNearlyEqual(points[i].X, expected[i].X) || !floatsNearlyEqual(points[i].Y, expected[i].Y) {
			t.Fatalf("point %d mismatch: got %+v, expected %+v", i, points[i], expected[i])
		}
	}
}

func TestApplyNamedStatePropertiesWithGeometry(t *testing.T) {
	shape := &components.Shape{Width: 50, Height: 40, X: 0, Y: 0}
	props := &stubProps{}
	node := parser.Node{
		State: "custom",
		States: map[string]parser.State{
			"custom": {Name: "custom", PropsDef: "x:12,y:24,w:300,h:150"},
		},
	}

	stateName := applyNamedStateProperties(node, shape, props, true)

	if stateName != "custom" {
		t.Fatalf("expected state name 'custom', got %q", stateName)
	}
	if !floatsNearlyEqual(shape.X, 12) || !floatsNearlyEqual(shape.Y, 24) {
		t.Fatalf("expected geometry translation to 12,24 got %f,%f", shape.X, shape.Y)
	}
	if !floatsNearlyEqual(shape.Width, 300) || !floatsNearlyEqual(shape.Height, 150) {
		t.Fatalf("expected geometry size 300x150 got %fx%f", shape.Width, shape.Height)
	}
	if props.last == "" {
		t.Fatalf("expected props parser to be invoked")
	}
}
