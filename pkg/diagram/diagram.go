package diagram

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/saasuke-labs/nagare/pkg/components"
	diagramapigateway "github.com/saasuke-labs/nagare/pkg/diagram/components/apigateway"
	diagramartifact "github.com/saasuke-labs/nagare/pkg/diagram/components/artifact"
	diagrambackgroundworker "github.com/saasuke-labs/nagare/pkg/diagram/components/backgroundworker"
	diagrambrowser "github.com/saasuke-labs/nagare/pkg/diagram/components/browser"
	diagramcdn "github.com/saasuke-labs/nagare/pkg/diagram/components/cdn"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/core"
	diagramcylinder "github.com/saasuke-labs/nagare/pkg/diagram/components/cylinder"
	diagramdatabase "github.com/saasuke-labs/nagare/pkg/diagram/components/database"
	diagramled "github.com/saasuke-labs/nagare/pkg/diagram/components/led"
	diagrammessagequeue "github.com/saasuke-labs/nagare/pkg/diagram/components/messagequeue"
	diagrampackage "github.com/saasuke-labs/nagare/pkg/diagram/components/packagecomponent"
	diagramrectangle "github.com/saasuke-labs/nagare/pkg/diagram/components/rectangle"
	diagramserver "github.com/saasuke-labs/nagare/pkg/diagram/components/server"
	diagramterminal "github.com/saasuke-labs/nagare/pkg/diagram/components/terminal"
	diagramvm "github.com/saasuke-labs/nagare/pkg/diagram/components/vm"
	"github.com/saasuke-labs/nagare/pkg/layout"
	"github.com/saasuke-labs/nagare/pkg/parser"
	nagareprops "github.com/saasuke-labs/nagare/pkg/props"
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
		Props:    nagareprops.ParseToMap(layoutPropsDef(d.AST)),
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
	nodeProps := make(map[string]any)
	rawDefs := make([]string, 0, 2)

	// Merge ID-based state (@db(...)) first.
	if s, ok := d.AST.Globals[node.Text]; ok {
		rawDefs = append(rawDefs, s.PropsDef)
		mergeProps(nodeProps, nagareprops.ParseToMap(s.PropsDef))
	}

	// Merge named state (@nginx(...)) referenced in declaration: db:Database@nginx
	if node.State != "" {
		if s, ok := d.AST.Globals[node.State]; ok {
			rawDefs = append(rawDefs, s.PropsDef)
			mergeProps(nodeProps, nagareprops.ParseToMap(s.PropsDef))
		}
	}

	if actions := collectActionsForNode(d.AST, node.Text); len(actions) > 0 {
		nodeProps["actions"] = actions
	}

	nodeProps["_rawProps"] = strings.Join(rawDefs, ",")

	// Include resolved geometry as numeric props when available.
	if shape, ok := d.Layout.NodeIndex[node.Text]; ok {
		if _, exists := nodeProps["x"]; !exists {
			nodeProps["x"] = shape.X
		}
		if _, exists := nodeProps["y"]; !exists {
			nodeProps["y"] = shape.Y
		}
		if _, exists := nodeProps["w"]; !exists {
			nodeProps["w"] = shape.Width
		}
		if _, exists := nodeProps["h"]; !exists {
			nodeProps["h"] = shape.Height
		}
	}

	renderNode := &RenderNode{
		Type:     strings.ToLower(string(node.Type)),
		ID:       node.Text,
		Props:    nodeProps,
		Children: make([]*RenderNode, 0, len(node.Children)),
	}

	for _, child := range node.Children {
		renderNode.Children = append(renderNode.Children, d.buildRenderNode(child))
	}

	return renderNode
}

func collectActionsForNode(ast parser.Node, nodeID string) map[string][]map[string]any {
	actions := make(map[string][]map[string]any)
	prefix := nodeID + "."

	states := ast.GlobalStates
	if len(states) == 0 {
		// Backward-compatible fallback if parser did not populate GlobalStates.
		states = make([]parser.State, 0, len(ast.Globals))
		for _, state := range ast.Globals {
			states = append(states, state)
		}
	}

	for _, state := range states {
		if !strings.HasPrefix(state.Name, prefix) {
			continue
		}
		actionName := strings.TrimPrefix(state.Name, prefix)
		if actionName == "" {
			continue
		}
		actions[actionName] = append(actions[actionName], nagareprops.ParseToMap(state.PropsDef))
	}

	return actions
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

// RenderTreeChildren recursively renders all descendants of a node.
// Recursion is centralized here to avoid circular dependencies in components.
func (d *Diagram) RenderTreeChildren(node *RenderNode) string {
	if node == nil {
		return ""
	}

	var b strings.Builder
	for _, child := range node.Children {
		b.WriteString(d.renderNodeRecursive(child, 0, 0))
		b.WriteString("\n")
	}
	return b.String()
}

func (d *Diagram) renderNodeRecursive(node *RenderNode, parentAbsX, parentAbsY float64) string {
	absShape := shapeFromNode(node)
	localShape := absShape
	localShape.X = absShape.X - parentAbsX
	localShape.Y = absShape.Y - parentAbsY

	current := d.drawNode(node, localShape)

	var children strings.Builder
	for _, child := range node.Children {
		children.WriteString(d.renderNodeRecursive(child, absShape.X, absShape.Y))
		children.WriteString("\n")
	}

	childrenContent := strings.TrimSpace(children.String())
	if strings.TrimSpace(current) == "" && childrenContent == "" {
		return ""
	}

	return fmt.Sprintf(`<g transform="translate(%.6f,%.6f)">%s%s</g>`, localShape.X, localShape.Y, current, children.String())
}

func (d *Diagram) drawNode(node *RenderNode, shape components.Shape) string {
	localProps := cloneProps(node.Props)
	// RenderTree recursion already wraps each node in a translated <g>, so
	// component templates must render at local origin to avoid applying X/Y twice.
	localProps["x"] = 0.0
	localProps["y"] = 0.0
	localProps["w"] = shape.Width
	localProps["h"] = shape.Height

	switch node.Type {
	case "browser":
		return diagrambrowser.DrawFromRenderNode(node.ID, localProps)
	case "vm":
		return diagramvm.DrawFromRenderNode(node.ID, localProps)
	case "server":
		return diagramserver.DrawFromRenderNode(node.ID, localProps)
	case "terminal":
		return diagramterminal.DrawFromRenderNode(node.ID, localProps)
	case "database":
		return diagramdatabase.DrawFromRenderNode(node.ID, localProps)
	case "cylinder":
		return diagramcylinder.DrawFromRenderNode(node.ID, localProps)
	case "led":
		return diagramled.DrawFromRenderNode(node.ID, localProps)
	case "messagequeue":
		return diagrammessagequeue.DrawFromRenderNode(node.ID, localProps)
	case "cdn":
		return diagramcdn.DrawFromRenderNode(node.ID, localProps)
	case "apigateway":
		return diagramapigateway.DrawFromRenderNode(node.ID, localProps)
	case "backgroundworker":
		return diagrambackgroundworker.DrawFromRenderNode(node.ID, localProps)
	case "package":
		return diagrampackage.DrawFromRenderNode(node.ID, localProps)
	case "artifact":
		return diagramartifact.DrawFromRenderNode(node.ID, localProps)
	case "rectangle":
		return diagramrectangle.DrawFromRenderNode(node.ID, localProps)
	default:
		return ""
	}
}

func cloneProps(src map[string]any) map[string]any {
	if src == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func shapeFromNode(node *RenderNode) components.Shape {
	return components.Shape{
		X:      core.FloatProp(node.Props, "x", 0),
		Y:      core.FloatProp(node.Props, "y", 0),
		Width:  core.FloatProp(node.Props, "w", 0),
		Height: core.FloatProp(node.Props, "h", 0),
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
		if k == "x" || k == "y" || k == "w" || k == "h" || k == "actions" || strings.HasPrefix(k, "_") {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s:%s", k, serializePropValue(v)))
	}

	return strings.Join(parts, ",")
}

func serializePropValue(v any) string {
	switch t := v.(type) {
	case string:
		if strings.HasPrefix(t, "&") {
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
// Internal helper: prefer calling pkg/nagare.RenderToSVG() from outside this module.
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

	return svg, nil
}

// CreateDiagramWithSize generates an SVG diagram and returns it along with
// the computed canvas size in pixels.
// Internal helper: prefer calling pkg/nagare.CreateDiagramWithSize() from outside this module.
func CreateDiagramWithSize(code string) (string, int, int, error) {
	diagram, err := ParseDiagram(code)
	if err != nil {
		return "", 0, 0, err
	}

	w, h := diagram.Layout.Bounds.Width, diagram.Layout.Bounds.Height
	renderTree := diagram.BuildRenderTree()
	componentsSVG := diagram.RenderTreeChildren(renderTree)
	componentsSVG += renderArrowComponents(diagram.Layout.Children)

	svg := fmt.Sprintf("<svg viewBox=\"0 0 %.6f %.6f\"  xmlns=\"http://www.w3.org/2000/svg\">\n"+
		"<!-- Background -->\n"+
		"<rect width=\"%.6f\" height=\"%.6f\" fill=\"#ccc\"></rect>\n"+
		"%s"+
		"</svg>", w, h, w, h, componentsSVG)

	return svg, normalizedCanvasDimension(w, 800), normalizedCanvasDimension(h, 400), nil
}

func normalizedCanvasDimension(calculated float64, fallback int) int {
	dim := int(math.Ceil(calculated))
	if dim <= 0 {
		dim = fallback
	}
	return dim
}
