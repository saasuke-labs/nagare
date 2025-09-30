package layout

import (
	"testing"

	"nagare/components"
	"nagare/parser"
)

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
		to := len(root.Children) + len(root.Connections)
		t.Fatalf("expected %d components, got %d", to, len(result.Children))
	}

	for i := 0; i < len(root.Connections); i++ {
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
