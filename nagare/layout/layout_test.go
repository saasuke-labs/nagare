package layout

import (
	"math"
	"nagare/components"
	"nagare/parser"
	"testing"
)

func almostEqual(a, b float64) bool {
	const epsilon = 1e-6
	return math.Abs(a-b) <= epsilon
}

func TestCalculateResolvesNestedServerAnchors(t *testing.T) {
	root := parser.Node{
		Children: []parser.Node{
			{
				Type: parser.NodeType("Browser"),
				Text: "browser1",
				States: map[string]parser.State{
					"browser1": {Name: "browser1", PropsDef: "x: 100, y: 50"},
				},
			},
			{
				Type: parser.NodeType("VM"),
				Text: "vm1",
				States: map[string]parser.State{
					"vm1": {Name: "vm1", PropsDef: "x: 200, y: 100"},
				},
				Children: []parser.Node{
					{
						Type: parser.NodeType("Server"),
						Text: "app",
						States: map[string]parser.State{
							"app": {Name: "app", PropsDef: "x: 25, y: 40"},
						},
					},
				},
			},
		},
		Connections: []parser.Connection{
			{
				From: parser.ConnectionEndpoint{NodeID: "app", Anchor: "w"},
				To:   parser.ConnectionEndpoint{NodeID: "browser1", Anchor: "e"},
			},
			{
				From: parser.ConnectionEndpoint{NodeID: "app", Anchor: "wn"},
				To:   parser.ConnectionEndpoint{NodeID: "browser1", Anchor: "en"},
			},
		},
	}

	result := Calculate(root, 960, 540)

	serverShape, ok := result.NodeIndex["app"]
	if !ok {
		t.Fatalf("expected server shape to be tracked in NodeIndex")
	}

	vmShape, ok := result.NodeIndex["vm1"]
	if !ok {
		t.Fatalf("expected vm shape to be tracked in NodeIndex")
	}

	expectedServerX := vmShape.X + vmShape.Width*components.VMContentAreaXRatio + 25
	expectedServerY := vmShape.Y + vmShape.Height*components.VMContentAreaYRatio + 40

	if !almostEqual(serverShape.X, expectedServerX) {
		t.Fatalf("server absolute X = %f, expected %f", serverShape.X, expectedServerX)
	}
	if !almostEqual(serverShape.Y, expectedServerY) {
		t.Fatalf("server absolute Y = %f, expected %f", serverShape.Y, expectedServerY)
	}

	if len(result.Connections) != 2 {
		t.Fatalf("expected 2 resolved connections, got %d", len(result.Connections))
	}

	browserShape := result.NodeIndex["browser1"]

	first := result.Connections[0]
	expectedFirstStart := Point{X: serverShape.X, Y: serverShape.Y + serverShape.Height*0.5}
	expectedFirstEnd := Point{X: browserShape.X + browserShape.Width, Y: browserShape.Y + browserShape.Height*0.5}
	if !almostEqual(first.Start.X, expectedFirstStart.X) || !almostEqual(first.Start.Y, expectedFirstStart.Y) {
		t.Fatalf("first connection start = (%f,%f), expected (%f,%f)", first.Start.X, first.Start.Y, expectedFirstStart.X, expectedFirstStart.Y)
	}
	if !almostEqual(first.End.X, expectedFirstEnd.X) || !almostEqual(first.End.Y, expectedFirstEnd.Y) {
		t.Fatalf("first connection end = (%f,%f), expected (%f,%f)", first.End.X, first.End.Y, expectedFirstEnd.X, expectedFirstEnd.Y)
	}

	second := result.Connections[1]
	expectedSecondStart := Point{X: serverShape.X, Y: serverShape.Y + serverShape.Height*0.25}
	expectedSecondEnd := Point{X: browserShape.X + browserShape.Width, Y: browserShape.Y + browserShape.Height*0.25}
	if !almostEqual(second.Start.X, expectedSecondStart.X) || !almostEqual(second.Start.Y, expectedSecondStart.Y) {
		t.Fatalf("second connection start = (%f,%f), expected (%f,%f)", second.Start.X, second.Start.Y, expectedSecondStart.X, expectedSecondStart.Y)
	}
	if !almostEqual(second.End.X, expectedSecondEnd.X) || !almostEqual(second.End.Y, expectedSecondEnd.Y) {
		t.Fatalf("second connection end = (%f,%f), expected (%f,%f)", second.End.X, second.End.Y, expectedSecondEnd.X, expectedSecondEnd.Y)
	}
}
