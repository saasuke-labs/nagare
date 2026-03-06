package diagram

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/saasuke-labs/nagare/pkg/components"
	diagramapigateway "github.com/saasuke-labs/nagare/pkg/diagram/components/apigateway"
	diagramartifact "github.com/saasuke-labs/nagare/pkg/diagram/components/artifact"
	diagrambackgroundworker "github.com/saasuke-labs/nagare/pkg/diagram/components/backgroundworker"
	diagrambrowser "github.com/saasuke-labs/nagare/pkg/diagram/components/browser"
	diagramcdn "github.com/saasuke-labs/nagare/pkg/diagram/components/cdn"
	diagramdatabase "github.com/saasuke-labs/nagare/pkg/diagram/components/database"
	diagrammessagequeue "github.com/saasuke-labs/nagare/pkg/diagram/components/messagequeue"
	diagrampackage "github.com/saasuke-labs/nagare/pkg/diagram/components/packagecomponent"
	diagramrectangle "github.com/saasuke-labs/nagare/pkg/diagram/components/rectangle"
	diagramserver "github.com/saasuke-labs/nagare/pkg/diagram/components/server"
	diagramterminal "github.com/saasuke-labs/nagare/pkg/diagram/components/terminal"
	diagramvm "github.com/saasuke-labs/nagare/pkg/diagram/components/vm"
	"github.com/saasuke-labs/nagare/pkg/layout"
	"github.com/saasuke-labs/nagare/pkg/parser"
	"github.com/saasuke-labs/nagare/pkg/tokenizer"
)

// Diagram represents a parsed and laid-out diagram structure
type Diagram struct {
	AST     parser.Node   // The parsed AST
	Layout  layout.Layout // The computed layout
	nodes   map[string]*DiagramNode
	astRoot parser.Node // Root AST node for reference
}

// DiagramNode represents a single component in the diagram with its metadata
type DiagramNode struct {
	ID    string            // Component ID from AST
	Type  string            // Component type (e.g., "server", "vm", "browser")
	Name  string            // Display name
	Shape components.Shape  // Geometry and position
	Props map[string]string // Component properties
}

// RenderNode is a tree-friendly structure for type-driven rendering.
// It intentionally stores props as a generic map to support mixed values
// (numbers, strings, alignment refs, percentages).
type RenderNode struct {
	Type     string
	ID       string
	Props    map[string]any
	Children []*RenderNode
}

// ToMap returns a map representation using the requested "_children" shape.
func (n *RenderNode) ToMap() map[string]any {
	children := make([]map[string]any, 0, len(n.Children))
	for _, child := range n.Children {
		children = append(children, child.ToMap())
	}

	out := map[string]any{
		"type":      n.Type,
		"id":        n.ID,
		"_children": children,
	}
	for k, v := range n.Props {
		if strings.HasPrefix(k, "_") {
			continue
		}
		out[k] = v
	}

	return out
}

// ParseDiagram parses DSL code and returns a navigable Diagram structure
func ParseDiagram(code string) (*Diagram, error) {
	tokens := tokenizer.Tokenize(code)
	ast, err := parser.Parse(tokens)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	const defaultCanvasWidth, defaultCanvasHeight = 800.0, 400.0
	l := layout.Calculate(ast, defaultCanvasWidth, defaultCanvasHeight)

	diagram := &Diagram{
		AST:     ast,
		Layout:  l,
		nodes:   make(map[string]*DiagramNode),
		astRoot: ast,
	}

	// Index nodes for easy lookup by building from AST
	indexNodes(ast, diagram.nodes, l.NodeIndex)

	return diagram, nil
}

// indexNodes recursively builds the node index from the AST
func indexNodes(node parser.Node, nodeMap map[string]*DiagramNode, shapeIndex map[string]components.Shape) {
	// Add the current node if it has an ID
	if node.Text != "" {
		shape, hasShape := shapeIndex[node.Text]
		if hasShape {
			nodeMap[node.Text] = &DiagramNode{
				ID:    node.Text,
				Type:  string(node.Type),
				Name:  node.Text,
				Shape: shape,
				Props: make(map[string]string),
			}
		}
	}

	// Recursively process children
	for _, child := range node.Children {
		indexNodes(child, nodeMap, shapeIndex)
	}
}

// GetComponent returns a component by its ID
func (d *Diagram) GetComponent(id string) *DiagramNode {
	return d.nodes[id]
}

// GetComponentsByType returns all components of a specific type
func (d *Diagram) GetComponentsByType(componentType string) []*DiagramNode {
	var result []*DiagramNode
	for _, node := range d.nodes {
		// Match by node type
		if node.Type == componentType {
			result = append(result, node)
		}
	}
	return result
}

// GetLayoutProperties returns the @layout properties from the diagram
func (d *Diagram) GetLayoutProperties() map[string]interface{} {
	layoutState, ok := d.AST.Globals["layout"]
	if !ok {
		return make(map[string]interface{})
	}
	// Return the raw props definition for now; can be parsed further if needed
	return map[string]interface{}{
		"propsDef": layoutState.PropsDef,
	}
}

// ValidateSingleLayout checks that there is exactly one layout component
func (d *Diagram) ValidateSingleLayout() error {
	layouts := d.GetComponentsByType("layout")
	if len(layouts) == 0 {
		return fmt.Errorf("diagram must have exactly one layout component, found 0")
	}
	if len(layouts) > 1 {
		return fmt.Errorf("diagram must have exactly one layout component, found %d", len(layouts))
	}
	return nil
}

// RootChildren returns the immediate children of the root AST node
func (d *Diagram) RootChildren() []parser.Node {
	return d.AST.Children
}

// BuildRenderTree builds a tree model that can be iterated by node type.
// Root node is the logical layout container with parsed layout props.
func (d *Diagram) BuildRenderTree() *RenderNode {
	root := &RenderNode{
		Type:     "layout",
		ID:       "layout",
		Props:    parsePropsDefSafe(layoutPropsDef(d.AST)),
		Children: make([]*RenderNode, 0, len(d.AST.Children)),
	}

	for _, child := range d.AST.Children {
		root.Children = append(root.Children, d.buildRenderNode(child))
	}

	return root
}

// BuildRenderTreeMap returns a map[string]any tree with "_children" nodes.
func (d *Diagram) BuildRenderTreeMap() map[string]any {
	return d.BuildRenderTree().ToMap()
}

func (d *Diagram) buildRenderNode(node parser.Node) *RenderNode {
	props := make(map[string]any)
	rawDefs := make([]string, 0, 2)

	// Merge ID-based state (@db(...)) first.
	if s, ok := d.AST.Globals[node.Text]; ok {
		rawDefs = append(rawDefs, s.PropsDef)
		mergeProps(props, parsePropsDefSafe(s.PropsDef))
	}

	// Merge named state (@nginx(...)) referenced in declaration: db:Database@nginx
	if node.State != "" {
		if s, ok := d.AST.Globals[node.State]; ok {
			rawDefs = append(rawDefs, s.PropsDef)
			mergeProps(props, parsePropsDefSafe(s.PropsDef))
		}
	}
	props["_rawProps"] = strings.Join(rawDefs, ",")

	// Include resolved geometry as numeric props when available.
	if shape, ok := d.Layout.NodeIndex[node.Text]; ok {
		if _, exists := props["x"]; !exists {
			props["x"] = shape.X
		}
		if _, exists := props["y"]; !exists {
			props["y"] = shape.Y
		}
		if _, exists := props["w"]; !exists {
			props["w"] = shape.Width
		}
		if _, exists := props["h"]; !exists {
			props["h"] = shape.Height
		}
	}

	renderNode := &RenderNode{
		Type:     strings.ToLower(string(node.Type)),
		ID:       node.Text,
		Props:    props,
		Children: make([]*RenderNode, 0, len(node.Children)),
	}

	for _, child := range node.Children {
		renderNode.Children = append(renderNode.Children, d.buildRenderNode(child))
	}

	return renderNode
}

func layoutPropsDef(ast parser.Node) string {
	if ast.Globals == nil {
		return ""
	}
	if layoutState, ok := ast.Globals["layout"]; ok {
		return layoutState.PropsDef
	}
	return ""
}

func mergeProps(dst map[string]any, src map[string]any) {
	for k, v := range src {
		dst[k] = v
	}
}

func parsePropsDefSafe(raw string) map[string]any {
	out := make(map[string]any)
	for _, part := range splitProps(raw) {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		if key == "" {
			continue
		}
		out[key] = coerceValue(strings.TrimSpace(kv[1]))
	}
	return out
}

func splitProps(raw string) []string {
	parts := []string{}
	if strings.TrimSpace(raw) == "" {
		return parts
	}

	var buf strings.Builder
	inQuotes := false
	quoteChar := byte(0)
	parenDepth := 0

	for i := 0; i < len(raw); i++ {
		c := raw[i]

		if (c == '"' || c == '\'') && (i == 0 || raw[i-1] != '\\') {
			if inQuotes && c == quoteChar {
				inQuotes = false
				quoteChar = 0
			} else if !inQuotes {
				inQuotes = true
				quoteChar = c
			}
			buf.WriteByte(c)
			continue
		}

		if !inQuotes {
			switch c {
			case '(':
				parenDepth++
			case ')':
				if parenDepth > 0 {
					parenDepth--
				}
			case ',':
				if parenDepth == 0 {
					part := strings.TrimSpace(buf.String())
					if part != "" {
						parts = append(parts, part)
					}
					buf.Reset()
					continue
				}
			}
		}

		buf.WriteByte(c)
	}

	if tail := strings.TrimSpace(buf.String()); tail != "" {
		parts = append(parts, tail)
	}

	return parts
}

func coerceValue(v string) any {
	v = strings.TrimSpace(v)
	v = strings.Trim(v, "\"")
	v = strings.Trim(v, "'")
	if v == "" {
		return ""
	}

	if n, err := strconv.ParseFloat(v, 64); err == nil {
		return n
	}

	// Keep percentages/alignment refs/symbolic strings as-is.
	return v
}

// RenderTreeChildren recursively renders all descendants of a node.
// Recursion is centralized here to avoid circular dependencies in components.
func (d *Diagram) RenderTreeChildren(node *RenderNode) string {
	if node == nil {
		return ""
	}

	rootWidth := floatProp(node.Props, "w", d.Layout.Bounds.Width)
	rootHeight := floatProp(node.Props, "h", d.Layout.Bounds.Height)

	var b strings.Builder
	for _, child := range node.Children {
		b.WriteString(d.renderNodeRecursive(child, 0, 0, rootWidth, rootHeight))
		b.WriteString("\n")
	}
	return b.String()
}

func (d *Diagram) renderNodeRecursive(node *RenderNode, parentAbsX, parentAbsY, parentWidth, parentHeight float64) string {
	absShape := shapeFromNode(node, parentWidth, parentHeight)
	localShape := absShape
	localShape.X = absShape.X - parentAbsX
	localShape.Y = absShape.Y - parentAbsY

	current := d.drawNode(node, localShape)

	var children strings.Builder
	for _, child := range node.Children {
		children.WriteString(d.renderNodeRecursive(child, absShape.X, absShape.Y, absShape.Width, absShape.Height))
		children.WriteString("\n")
	}

	childrenContent := strings.TrimSpace(children.String())
	if strings.TrimSpace(current) == "" && childrenContent == "" {
		return ""
	}

	return fmt.Sprintf(`<g transform="translate(%.6f,%.6f)">%s%s</g>`, localShape.X, localShape.Y, current, children.String())
}

func (d *Diagram) drawNode(node *RenderNode, shape components.Shape) string {
	raw := rawPropsFromNode(node)
	localShape := shape
	localShape.X = 0
	localShape.Y = 0

	switch node.Type {
	case "browser":
		comp := diagrambrowser.NewLegacy(node.ID)
		comp.Shape = localShape
		_ = comp.Props.Parse(raw)
		return comp.Draw()
	case "vm":
		comp := diagramvm.NewLegacy(node.ID)
		comp.Shape = localShape
		_ = comp.Props.Parse(raw)
		return comp.Draw()
	case "server":
		comp := diagramserver.NewLegacy(node.ID)
		comp.SetShape(localShape)
		_ = comp.Comp.ApplyProps(raw)
		return comp.Draw()
	case "terminal":
		comp := diagramterminal.NewLegacy(node.ID)
		comp.SetShape(localShape)
		_ = comp.Comp.ApplyProps(raw)
		return comp.Draw()
	case "database":
		comp := diagramdatabase.NewLegacy(node.ID)
		comp.Shape = localShape
		_ = comp.Props.Parse(raw)
		return comp.Draw()
	case "messagequeue":
		comp := diagrammessagequeue.NewLegacy(node.ID)
		comp.Shape = localShape
		_ = comp.Props.Parse(raw)
		return comp.Draw()
	case "cdn":
		comp := diagramcdn.NewLegacy(node.ID)
		comp.Shape = localShape
		_ = comp.Props.Parse(raw)
		return comp.Draw()
	case "apigateway":
		comp := diagramapigateway.NewLegacy(node.ID)
		comp.Shape = localShape
		_ = comp.Props.Parse(raw)
		return comp.Draw()
	case "backgroundworker":
		comp := diagrambackgroundworker.NewLegacy(node.ID)
		comp.Shape = localShape
		_ = comp.Props.Parse(raw)
		return comp.Draw()
	case "package":
		comp := diagrampackage.NewLegacy(node.ID)
		comp.Shape = localShape
		_ = comp.Props.Parse(raw)
		return comp.Draw()
	case "artifact":
		comp := diagramartifact.NewLegacy(node.ID)
		comp.Shape = localShape
		_ = comp.Props.Parse(raw)
		return comp.Draw()
	case "rectangle":
		comp := diagramrectangle.NewLegacy(node.ID)
		comp.SetShape(localShape)
		_ = comp.Comp.ApplyProps(raw)
		return comp.Draw()
	default:
		return ""
	}
}

func shapeFromNode(node *RenderNode, parentWidth, parentHeight float64) components.Shape {
	return components.Shape{
		X:      floatProp(node.Props, "x", parentWidth),
		Y:      floatProp(node.Props, "y", parentHeight),
		Width:  floatProp(node.Props, "w", parentWidth),
		Height: floatProp(node.Props, "h", parentHeight),
	}
}

func rawPropsFromNode(node *RenderNode) string {
	if v, ok := node.Props["_rawProps"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}

	parts := make([]string, 0, len(node.Props))
	for k, v := range node.Props {
		if k == "x" || k == "y" || k == "w" || k == "h" || strings.HasPrefix(k, "_") {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s:%s", k, serializePropValue(v)))
	}

	return strings.Join(parts, ",")
}

func serializePropValue(v any) string {
	switch t := v.(type) {
	case string:
		if strings.HasPrefix(t, "&") || strings.HasSuffix(t, "%") {
			return t
		}
		return fmt.Sprintf("\"%s\"", t)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func floatProp(props map[string]any, key string, base float64) float64 {
	v, ok := props[key]
	if !ok {
		return 0
	}

	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case string:
		s := strings.TrimSpace(t)
		if strings.HasSuffix(s, "%") {
			pct, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(s, "%")), 64)
			if err != nil {
				return 0
			}
			if base <= 0 {
				return pct
			}
			return base * (pct / 100.0)
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}

func renderArrowComponents(children []components.Component) string {
	var b strings.Builder
	for _, c := range children {
		if arrow, ok := c.(*components.Arrow); ok {
			b.WriteString(arrow.Draw())
			b.WriteString("\n")
		}
	}
	return b.String()
}

// CreateDiagram generates an SVG diagram from the provided code and returns it as a string.
func CreateDiagram(code string) (string, error) {
	diagram, err := ParseDiagram(code)
	if err != nil {
		return "", err
	}

	w, h := diagram.Layout.Bounds.Width, diagram.Layout.Bounds.Height

	renderTree := diagram.BuildRenderTree()
	componentsSVG := diagram.RenderTreeChildren(renderTree)
	componentsSVG += renderArrowComponents(diagram.Layout.Children)

	// For now, just return a placeholder SVG
	// In the future, this will use the diagram structure to render properly
	svg := fmt.Sprintf("<svg viewBox=\"0 0 %.6f %.6f\"  xmlns=\"http://www.w3.org/2000/svg\">\n"+
		"<!-- Background -->\n"+
		"<rect width=\"%.6f\" height=\"%.6f\" fill=\"#ccc\"></rect>\n"+
		"%s"+
		"</svg>", w, h, w, h, componentsSVG)

	fmt.Println(svg)

	return svg, nil
}

// // CreateDiagramWithSize generates an SVG diagram and returns the SVG along with the computed canvas size.
// func CreateDiagramWithSize(code string) (string, int, int, error) {
// 	return RenderToSVGWithSize(code)
// }

// // RenderToSVG renders diagram DSL input into SVG.
// func RenderToSVG(code string) (string, error) {
// 	svg, _, _, err := RenderToSVGWithSize(code)
// 	return svg, err
// }

// // RenderToSVGWithSize renders a diagram and returns SVG with computed canvas size.
// func RenderToSVGWithSize(code string) (string, int, int, error) {
// 	tokens := tokenizer.Tokenize(code)
// 	ast, err := parser.Parse(tokens)
// 	if err != nil {
// 		return "", 0, 0, fmt.Errorf("parse error: %w", err)
// 	}

// 	const defaultCanvasWidth, defaultCanvasHeight = 800.0, 400.0
// 	l := layout.Calculate(ast, defaultCanvasWidth, defaultCanvasHeight)

// 	canvasWidth := normalizedCanvasDimension(l.Bounds.Width, defaultCanvasWidth)
// 	canvasHeight := normalizedCanvasDimension(l.Bounds.Height, defaultCanvasHeight)
// 	svg := renderer.Render(l, canvasWidth, canvasHeight)

// 	return svg, canvasWidth, canvasHeight, nil
// }

// // RenderToSVGWithDebug renders a diagram and includes parser/layout debug output.
// func RenderToSVGWithDebug(code string) (string, error) {
// 	fmt.Printf("Input code:\n%s\n", code)

// 	tokens := tokenizer.Tokenize(code)
// 	fmt.Printf("Tokens: %+v\n", tokens)

// 	ast, err := parser.Parse(tokens)
// 	if err != nil {
// 		return "", fmt.Errorf("parse error: %w", err)
// 	}

// 	fmt.Printf("AST: \n%+v\n", ast)

// 	const defaultCanvasWidth, defaultCanvasHeight = 800.0, 400.0
// 	l := layout.Calculate(ast, defaultCanvasWidth, defaultCanvasHeight)
// 	fmt.Printf("Layout: \n%+v\n", l)

// 	canvasWidth := normalizedCanvasDimension(l.Bounds.Width, defaultCanvasWidth)
// 	canvasHeight := normalizedCanvasDimension(l.Bounds.Height, defaultCanvasHeight)
// 	svg := renderer.Render(l, canvasWidth, canvasHeight)
// 	fmt.Println(svg)

// 	return svg, nil
// }

// func normalizedCanvasDimension(calculated, fallback float64) int {
// 	dim := int(calculated)
// 	if dim == 0 {
// 		dim = int(fallback)
// 	}
// 	return dim
// }
