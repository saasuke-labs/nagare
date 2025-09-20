package mermaid

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Node struct {
	ID    string
	Label string
	Type  string // Browser|Server|DB|VM|Generic
	// Hierarchy: if Parent != "", this node belongs to that subgraph
	Parent string
}

type Edge struct {
	From, To string
	Label    string
}

type Graph struct {
	Direction string // RIGHT|DOWN|LEFT|UP
	Nodes     map[string]*Node
	Edges     []Edge
	// Subgraph id -> label
	Subgraphs map[string]string
}

func NewGraph() *Graph {
	return &Graph{
		Direction: "RIGHT",
		Nodes:     map[string]*Node{},
		Edges:     []Edge{},
		Subgraphs: map[string]string{},
	}
}

var (
	reFlowHdr  = regexp.MustCompile(`^flowchart\s+(LR|RL|TB|BT)\b`)
	reSubStart = regexp.MustCompile(`^subgraph\s+([A-Za-z0-9_]+)(?:\$begin:math:display\$(.*?)\$end:math:display\$)?`)
	reSubEnd   = regexp.MustCompile(`^end$`)
	// Node forms: ID[label] or ID((label)) or ID([label]) optionally :::Type
	reNode = regexp.MustCompile(`^([A-Za-z0-9_]+)\s*(?:\$begin:math:display\$(.*?)\$end:math:display\$|\$begin:math:text\$\((.*?)\$end:math:text\$|\$begin:math:text\$\[(.*?)\]\$end:math:text\$)?(?:::([A-Za-z0-9_]+))?$`)
	// Edge: A --> B : optional label
	reEdge = regexp.MustCompile(`^([A-Za-z0-9_]+)\s*-{1,}>\s*([A-Za-z0-9_]+)(?:\s*:\s*(.*))?$`)
)

func ParseFlowchart(src string) (*Graph, error) {
	g := NewGraph()
	stack := []string{} // subgraph id stack; empty means root

	sc := bufio.NewScanner(strings.NewReader(src))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "%%") {
			continue
		}
		if m := reFlowHdr.FindStringSubmatch(line); m != nil {
			dir := map[string]string{"LR": "RIGHT", "RL": "LEFT", "TB": "DOWN", "BT": "UP"}[strings.ToUpper(m[1])]
			if dir != "" {
				g.Direction = dir
			}
			continue
		}
		if m := reSubStart.FindStringSubmatch(line); m != nil {
			id := m[1]
			label := m[2]
			if label == "" {
				label = id
			}
			g.Subgraphs[id] = label
			stack = append(stack, id)
			continue
		}
		if reSubEnd.MatchString(line) {
			if len(stack) == 0 {
				return nil, fmt.Errorf("unbalanced 'end'")
			}
			stack = stack[:len(stack)-1]
			continue
		}
		if m := reEdge.FindStringSubmatch(line); m != nil {
			g.ensureNode(m[1], m[1], top(stack))
			g.ensureNode(m[2], m[2], top(stack))
			lbl := strings.TrimSpace(m[3])
			g.Edges = append(g.Edges, Edge{From: m[1], To: m[2], Label: lbl})
			continue
		}
		if m := reNode.FindStringSubmatch(line); m != nil {
			id := m[1]
			lbl := firstNonEmpty(m[2], m[3], m[4], id)
			tp := m[5]
			if tp == "" {
				tp = "Generic"
			}
			n := g.ensureNode(id, lbl, top(stack))
			n.Type = tp
			continue
		}
		// ignore: classDef/style/others for now
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return g, nil
}

func (g *Graph) ensureNode(id, label, parent string) *Node {
	if n, ok := g.Nodes[id]; ok {
		// keep earlier label if present
		if n.Label == "" {
			n.Label = label
		}
		return n
	}
	n := &Node{ID: id, Label: label, Type: "Generic", Parent: parent}
	g.Nodes[id] = n
	return n
}

func top(stack []string) string {
	if len(stack) == 0 {
		return ""
	}
	return stack[len(stack)-1]
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

func (g *Graph) ToDOT() string {
	var b strings.Builder
	// LR/TB mapping for dot
	dir := map[string]string{"RIGHT": "LR", "LEFT": "RL", "DOWN": "TB", "UP": "BT"}[g.Direction]
	if dir == "" {
		dir = "LR"
	}

	b.WriteString("digraph G {\n")
	b.WriteString(`rankdir=` + dir + ";\n")
	b.WriteString(`node [shape=rect, style="rounded"];` + "\n")
	b.WriteString(`graph [splines=ortho];` + "\n")
	// Subgraphs as clusters
	// group nodes by parent
	children := map[string][]*Node{}
	for _, n := range g.Nodes {
		children[n.Parent] = append(children[n.Parent], n)
	}
	// emit root nodes
	for _, n := range children[""] {
		fmt.Fprintf(&b, "%s [label=%q];\n", n.ID, n.Label)
	}
	// emit clusters
	for id, label := range g.Subgraphs {
		fmt.Fprintf(&b, "subgraph cluster_%s {\n", id)
		fmt.Fprintf(&b, "label=%q;\n", label)
		for _, n := range children[id] {
			fmt.Fprintf(&b, "%s [label=%q];\n", n.ID, n.Label)
		}
		b.WriteString("}\n")
	}
	// edges
	for i, e := range g.Edges {
		if e.Label != "" {
			fmt.Fprintf(&b, "e%d: %s -> %s [label=%q];\n", i, e.From, e.To, e.Label)
		} else {
			fmt.Fprintf(&b, "%s -> %s;\n", e.From, e.To)
		}
	}
	b.WriteString("}\n")
	return b.String()
}

type Box struct{ X, Y, W, H float64 }   // center-based (Graphviz uses points)
type Poly struct{ Points [][2]float64 } // edge polyline

type Layout struct {
	Nodes map[string]Box
	Edges map[[2]string]Poly
	// canvas size in points (inch*72)
	W, H float64
}

func LayoutWithDotPlain(dot string) (*Layout, error) {
	cmd := exec.Command("dot", "-Tplain")
	cmd.Stdin = strings.NewReader(dot)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("dot: %w", err)
	}

	L := &Layout{Nodes: map[string]Box{}, Edges: map[[2]string]Poly{}}
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	var scale float64 = 72.0 // points per inch
	for sc.Scan() {
		parts := strings.Fields(sc.Text())
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {
		case "graph":
			// graph width height
			if len(parts) >= 3 {
				w, _ := strconv.ParseFloat(parts[1], 64)
				h, _ := strconv.ParseFloat(parts[2], 64)
				L.W, L.H = w*scale, h*scale
			}
		case "node":
			// node name x y w h label style shape color fillcolor
			if len(parts) >= 6 {
				id := parts[1]
				x, _ := strconv.ParseFloat(parts[2], 64)
				y, _ := strconv.ParseFloat(parts[3], 64)
				w, _ := strconv.ParseFloat(parts[4], 64)
				h, _ := strconv.ParseFloat(parts[5], 64)
				L.Nodes[id] = Box{X: x * scale, Y: (L.H - y*scale), W: w * scale, H: h * scale}
			}
		case "edge":
			// edge tail head n x1 y1 ... xn yn
			if len(parts) >= 5 {
				from := parts[1]
				to := parts[2]
				npts, _ := strconv.Atoi(parts[3])
				ps := make([][2]float64, 0, npts)
				for i := 0; i < npts; i++ {
					px, _ := strconv.ParseFloat(parts[4+2*i], 64)
					py, _ := strconv.ParseFloat(parts[5+2*i], 64)
					ps = append(ps, [2]float64{px * scale, (L.H - py*scale)})
				}
				L.Edges[[2]string{from, to}] = Poly{Points: ps}
			}
		}
	}
	return L, nil
}
