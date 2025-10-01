package renderer

import (
	"strings"
	"testing"

	"github.com/saasuke-labs/nagare/pkg/layout"
	"github.com/saasuke-labs/nagare/pkg/parser"
)

func TestRenderIncludesStraightArrow(t *testing.T) {
	root := parser.Node{
		Globals: make(map[string]parser.State),
		Children: []parser.Node{
			{
				Type: "Browser",
				Text: "left",
				States: map[string]parser.State{
					"left": {Name: "left", PropsDef: "x:100,y:100,w:200,h:120"},
				},
			},
			{
				Type: "Browser",
				Text: "right",
				States: map[string]parser.State{
					"right": {Name: "right", PropsDef: "x:500,y:100,w:200,h:120"},
				},
			},
		},
		Connections: []parser.Connection{
			{
				FromID: "left",
				FromAnchor: parser.AnchorDescriptor{
					Raw:        "e",
					Horizontal: 1,
				},
				ToID: "right",
				ToAnchor: parser.AnchorDescriptor{
					Raw:        "w",
					Horizontal: -1,
				},
			},
		},
	}

	l := layout.Calculate(root, 1000, 600)
	svg := Render(l, 1000, 600)

	if !strings.Contains(svg, "points=\"300.00,160.00 500.00,160.00\"") {
		t.Fatalf("expected straight arrow polyline, got: %s", svg)
	}
}

func TestRenderIncludesElbowArrow(t *testing.T) {
	root := parser.Node{
		Globals: make(map[string]parser.State),
		Children: []parser.Node{
			{
				Type: "Browser",
				Text: "northBox",
				States: map[string]parser.State{
					"northBox": {Name: "northBox", PropsDef: "x:300,y:400,w:200,h:100"},
				},
			},
			{
				Type: "Browser",
				Text: "eastBox",
				States: map[string]parser.State{
					"eastBox": {Name: "eastBox", PropsDef: "x:600,y:200,w:200,h:100"},
				},
			},
		},
		Connections: []parser.Connection{
			{
				FromID: "northBox",
				FromAnchor: parser.AnchorDescriptor{
					Raw:      "n",
					Vertical: -1,
				},
				ToID: "eastBox",
				ToAnchor: parser.AnchorDescriptor{
					Raw:        "e",
					Horizontal: 1,
				},
			},
		},
	}

	l := layout.Calculate(root, 1200, 800)
	svg := Render(l, 1200, 800)

	expectedPoints := "points=\"400.00,400.00 400.00,376.00 800.00,376.00 800.00,250.00\""
	if !strings.Contains(svg, expectedPoints) {
		t.Fatalf("expected elbow arrow polyline %s, got: %s", expectedPoints, svg)
	}
}
