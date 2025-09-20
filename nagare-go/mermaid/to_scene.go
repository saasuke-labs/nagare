// file: mermaid/to_scene.go
package mermaid

import (
	"fmt"
	"nagare-go/model"
)

// type SceneEdge struct {
// 	ID     string `json:"id"`
// 	From   string `json:"from"`
// 	To     string `json:"to"`
// 	Label  string `json:"label,omitempty"`
// 	FlowOn []KfN  `json:"flowOn,omitempty"`
// }

// type Scene struct {
// 	Width, Height int         `json:"width,omitempty"`
// 	FPS           float64     `json:"fps,omitempty"`
// 	DurationSec   float64     `json:"durationSec"`
// 	Bg, Fg        string      `json:"bg,omitempty"`
// 	Nodes         []SceneNode `json:"nodes"`
// 	Edges         []SceneEdge `json:"edges"`
// }

func BuildScene(g *Graph, L *Layout, fps, duration float64, w, h int) model.Scene {
	nodes := make([]model.SceneNode, 0, len(g.Nodes))
	for id, n := range g.Nodes {
		box := L.Nodes[id]
		// convert center-based box → top-left x,y for your renderer
		x := box.X - box.W/2
		y := box.Y - box.H/2
		nodes = append(nodes, model.SceneNode{
			ID: id, Label: n.Label, X: x, Y: y, W: box.W, H: box.H,
			// simple default: fade in at 0.2→0.5s
			Opacity: []model.KfN{{T: 0.2, V: 0}, {T: 0.5, V: 1}},
		})
	}
	edges := make([]model.SceneEdge, 0, len(g.Edges))
	for i, e := range g.Edges {
		edges = append(edges, model.SceneEdge{
			ID: fmt.Sprintf("e%d", i), From: e.From, To: e.To, Label: e.Label,
			FlowOn: []model.KfN{{T: 0.8, V: 0}, {T: 1.0, V: 1}},
		})
	}
	return model.Scene{
		Width: w, Height: h, FPS: fps, DurationSec: duration,
		Bg: "#ffffff", Fg: "#222222",
		Nodes: nodes, Edges: edges,
	}
}
