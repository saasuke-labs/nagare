package layout

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
	"github.com/saasuke-labs/nagare/pkg/parser"
	"github.com/saasuke-labs/nagare/pkg/props"
)

const (
	defaultBrowserWidth      = 640.0
	defaultBrowserHeight     = 420.0
	defaultVMWidth           = 640.0
	defaultVMHeight          = 420.0
	defaultServerWidth       = 200.0
	defaultServerHeight      = 140.0
	defaultTerminalWidth     = 360.0
	defaultTerminalHeight    = 220.0
	defaultDatabaseWidth     = 200.0
	defaultDatabaseHeight    = 200.0
	defaultCylinderWidth     = 200.0
	defaultCylinderHeight    = 200.0
	defaultLedWidth          = 20.0
	defaultLedHeight         = 20.0
	defaultQueueWidth        = 220.0
	defaultQueueHeight       = 180.0
	defaultCDNWidth          = 200.0
	defaultCDNHeight         = 160.0
	defaultAPIGatewayWidth   = 220.0
	defaultAPIGatewayHeight  = 180.0
	defaultBackgroundWorkerW = 220.0
	defaultBackgroundWorkerH = 180.0
	defaultPackageWidth      = 200.0
	defaultPackageHeight     = 180.0
	defaultArtifactWidth     = 200.0
	defaultArtifactHeight    = 180.0
	defaultComponentX        = 0.0
	defaultComponentY        = 0.0
	arrowElbowPadding        = 24.0
	floatEqualityEpsilon     = 0.0001
)

const (
	componentTypeBrowser          = "Browser"
	componentTypeVM               = "VM"
	componentTypeServer           = "Server"
	componentTypeRectangle        = "Rectangle"
	componentTypeTerminal         = "Terminal"
	componentTypeDatabase         = "Database"
	componentTypeCylinder         = "Cylinder"
	componentTypeLed              = "Led"
	componentTypeMessageQueue     = "MessageQueue"
	componentTypeQueue            = "Queue"
	componentTypeCDN              = "CDN"
	componentTypeEdge             = "Edge"
	componentTypeAPIGateway       = "APIGateway"
	componentTypeBackgroundWorker = "BackgroundWorker"
	componentTypePackage          = "Package"
	componentTypeArtifact         = "Artifact"
	componentTypeFile             = "File"
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

// Arrow contains the resolved geometry for a parsed connection.
type Arrow struct {
	FromID      string
	ToID        string
	FromAnchor  string
	ToAnchor    string
	Start       core.Point
	End         core.Point
	BendPoints  []core.Point
	Style       string
	MarkerStart bool
	MarkerEnd   bool
}

type geometryProps struct {
	X      interface{} `prop:"x"`
	Y      interface{} `prop:"y"`
	Width  interface{} `prop:"w"`
	Height interface{} `prop:"h"`
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

func applyGeometryProps(shape *components.Shape, geom geometryProps, nodeIndex map[string]components.Shape) {
	if w, ok := parseNumericValue(geom.Width); ok {
		shape.Width = w
	}
	if h, ok := parseNumericValue(geom.Height); ok {
		shape.Height = h
	}
	if geom.X != nil {
		if intVal, ok := geom.X.(int); ok {
			shape.X = float64(intVal)
		} else if strVal, ok := geom.X.(string); ok && strings.HasPrefix(strVal, "&") {
			// Handle alignment reference - store for later resolution
			if shape.AlignmentRefs == nil {
				shape.AlignmentRefs = make(map[string]string)
			}
			shape.AlignmentRefs["x"] = strVal
		} else if strVal, ok := geom.X.(string); ok {
			// Check if this looks like a broken alignment reference (e.g., "browser  c")
			if strings.Contains(strVal, "  ") {
				if shape.AlignmentRefs == nil {
					shape.AlignmentRefs = make(map[string]string)
				}
				shape.AlignmentRefs["x_string"] = strVal
			} else if x, ok := parseNumericValue(strVal); ok {
				shape.X = x
			}
		} else if x, ok := parseNumericValue(geom.X); ok {
			shape.X = x
		}
	}
	if geom.Y != nil {
		if intVal, ok := geom.Y.(int); ok {
			shape.Y = float64(intVal)
		} else if strVal, ok := geom.Y.(string); ok && strings.HasPrefix(strVal, "&") {
			// Handle alignment reference - store for later resolution
			if shape.AlignmentRefs == nil {
				shape.AlignmentRefs = make(map[string]string)
			}
			shape.AlignmentRefs["y"] = strVal
		} else if strVal, ok := geom.Y.(string); ok {
			// Check if this looks like a broken alignment reference (e.g., "browser  c")
			if strings.Contains(strVal, "  ") {
				if shape.AlignmentRefs == nil {
					shape.AlignmentRefs = make(map[string]string)
				}
				shape.AlignmentRefs["y_string"] = strVal
			} else if y, ok := parseNumericValue(strVal); ok {
				shape.Y = y
			}
		} else if y, ok := parseNumericValue(geom.Y); ok {
			shape.Y = y
		}
	}
}

func parseNumericValue(v interface{}) (float64, bool) {
	if v == nil {
		return 0, false
	}

	switch t := v.(type) {
	case int:
		return float64(t), true
	case float64:
		return t, true
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return 0, false
		}
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, false
		}
		return n, true
	default:
		return 0, false
	}
}

func resolveAlignmentReference(ref string, nodeIndex map[string]components.Shape, currentShape *components.Shape) (float64, error) {
	// Parse &component.alignment syntax
	if !strings.HasPrefix(ref, "&") {
		return 0, fmt.Errorf("alignment reference must start with &")
	}

	// Remove & prefix
	ref = strings.TrimPrefix(ref, "&")

	// Split by dot
	parts := strings.Split(ref, ".")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid alignment reference format: expected &component.alignment")
	}

	componentName := parts[0]
	alignment := parts[1]

	// Find the target component
	targetShape, exists := nodeIndex[componentName]
	if !exists {
		return 0, fmt.Errorf("component %s not found", componentName)
	}

	// Calculate alignment position
	switch alignment {
	case "c": // center
		return targetShape.Y + targetShape.Height/2 - currentShape.Height/2, nil
	case "t": // top
		return targetShape.Y, nil
	case "b": // bottom
		return targetShape.Y + targetShape.Height - currentShape.Height, nil
	default:
		return 0, fmt.Errorf("unsupported alignment: %s", alignment)
	}
}

func resolveAlignmentReferences(nodeIndex map[string]components.Shape) {
	// Iterate through all shapes and resolve alignment references
	for componentName, shape := range nodeIndex {
		updated := false

		// Check if there are alignment references to resolve
		if shape.AlignmentRefs != nil {
			for axis, ref := range shape.AlignmentRefs {
				resolved, err := resolveAlignmentReference(ref, nodeIndex, &shape)
				if err != nil {
					// silently ignore unresolvable alignment references
					continue
				}

				switch axis {
				case "x":
					shape.X = resolved
					updated = true
				case "y":
					shape.Y = resolved
					updated = true
				}
			}
		}

		// Also check for alignment patterns in string values (fallback for current parsing)
		// This handles the case where tokenizer breaks "&browser.c" into "browser  c"
		if strY, ok := shape.AlignmentRefs["y_string"]; ok {
			// Try to reconstruct the alignment reference
			reconstructed := strings.ReplaceAll(strY, "  ", ".")
			reconstructed = "&" + reconstructed

			resolved, err := resolveAlignmentReference(reconstructed, nodeIndex, &shape)
			if err != nil {
				// silently ignore unresolvable reconstructed alignment references
			} else {
				shape.Y = resolved
				updated = true
			}
		}

		// Update the nodeIndex with the modified shape
		if updated {
			nodeIndex[componentName] = shape
		}
	}
}

func syncComponentGeometry(children []components.Component, nodeIndex map[string]components.Shape) {
	for _, child := range children {
		switch comp := child.(type) {
		case *diagrambrowser.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
			}
		case *components.Server:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
			}
		case *diagramserver.Adapter:
			if shape, ok := nodeIndex[comp.Text]; ok {
				comp.SetShape(shape)
			}
		case *components.Terminal:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
			}
		case *diagramterminal.Adapter:
			if shape, ok := nodeIndex[comp.Text]; ok {
				comp.SetShape(shape)
			}
		case *diagramdatabase.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
			}
		case *diagramcylinder.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
			}
		case *diagramled.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
			}
		case *diagrammessagequeue.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
			}
		case *diagramcdn.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
			}
		case *diagramapigateway.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
			}
		case *diagrambackgroundworker.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
			}
		case *diagrampackage.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
			}
		case *diagramartifact.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
			}
		case *diagramvm.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
			}
			syncVMChildGeometry(comp, nodeIndex)
		case *components.Rectangle:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
			}
		case *diagramrectangle.Adapter:
			if shape, ok := nodeIndex[comp.Text]; ok {
				comp.SetShape(shape)
			}
		}
	}
}

func applyResolvedShape(target *components.Shape, resolved components.Shape) {
	if target == nil {
		return
	}
	target.X = resolved.X
	target.Y = resolved.Y
	target.Width = resolved.Width
	target.Height = resolved.Height
}

func syncVMChildGeometry(vm *diagramvm.Component, nodeIndex map[string]components.Shape) {
	contentOriginX := vm.Shape.X + vm.Shape.Width*diagramvm.VMContentAreaXRatio
	contentOriginY := vm.Shape.Y + vm.Shape.Height*diagramvm.VMContentAreaYRatio

	for _, child := range vm.Children {
		switch comp := child.(type) {
		case *components.Server:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
				comp.X -= contentOriginX
				comp.Y -= contentOriginY
			}
		case *diagramserver.Adapter:
			if shape, ok := nodeIndex[comp.Text]; ok {
				comp.SetShape(shape)
				comp.Offset(contentOriginX, contentOriginY)
			}
		case *components.Terminal:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
				comp.X -= contentOriginX
				comp.Y -= contentOriginY
			}
		case *diagramterminal.Adapter:
			if shape, ok := nodeIndex[comp.Text]; ok {
				comp.SetShape(shape)
				comp.Offset(contentOriginX, contentOriginY)
			}
		case *diagramdatabase.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
				comp.X -= contentOriginX
				comp.Y -= contentOriginY
			}
		case *diagramcylinder.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
				comp.X -= contentOriginX
				comp.Y -= contentOriginY
			}
		case *diagramled.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
				comp.X -= contentOriginX
				comp.Y -= contentOriginY
			}
		case *diagrammessagequeue.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
				comp.X -= contentOriginX
				comp.Y -= contentOriginY
			}
		case *diagramcdn.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
				comp.X -= contentOriginX
				comp.Y -= contentOriginY
			}
		case *diagramapigateway.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
				comp.X -= contentOriginX
				comp.Y -= contentOriginY
			}
		case *diagrambackgroundworker.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
				comp.X -= contentOriginX
				comp.Y -= contentOriginY
			}
		case *diagrampackage.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
				comp.X -= contentOriginX
				comp.Y -= contentOriginY
			}
		case *diagramartifact.Component:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
				comp.X -= contentOriginX
				comp.Y -= contentOriginY
			}
		case *components.Rectangle:
			if shape, ok := nodeIndex[comp.Text]; ok {
				applyResolvedShape(&comp.Shape, shape)
				comp.X -= contentOriginX
				comp.Y -= contentOriginY
			}
		case *diagramrectangle.Adapter:
			if shape, ok := nodeIndex[comp.Text]; ok {
				comp.SetShape(shape)
				comp.Offset(contentOriginX, contentOriginY)
			}
		}
	}
}

// Calculate computes the layout for an AST
func Calculate(node parser.Node, canvasWidth, canvasHeight float64) (Layout, error) {
	boundsWidth, boundsHeight := calculateCanvasBounds(node, canvasWidth, canvasHeight)
	nodeIndex := make(map[string]components.Shape)

	children := make([]components.Component, 0, len(node.Children))
	for _, child := range node.Children {
		children = append(children, buildComponentTree(child, nodeIndex)...)
	}

	// Resolve alignment references after all components are positioned
	resolveAlignmentReferences(nodeIndex)
	syncComponentGeometry(children, nodeIndex)

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
	}, nil
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
		return boundsWidth, boundsHeight
	}

	if w, ok := parseNumericValue(geometry.Width); ok {
		boundsWidth = w
	}
	if h, ok := parseNumericValue(geometry.Height); ok {
		boundsHeight = h
	}

	return boundsWidth, boundsHeight
}

func buildComponentTree(node parser.Node, nodeIndex map[string]components.Shape) []components.Component {
	switch string(node.Type) {
	case componentTypeBrowser:
		return []components.Component{buildBrowser(node, nodeIndex)}
	case componentTypeVM:
		return []components.Component{buildVM(node, nodeIndex)}
	case componentTypeServer:
		return []components.Component{buildServer(node, nil, nodeIndex)}
	case componentTypeTerminal:
		return []components.Component{buildTerminal(node, nil, nodeIndex)}
	case componentTypeDatabase:
		return []components.Component{buildDatabase(node, nil, nodeIndex)}
	case componentTypeCylinder:
		return []components.Component{buildCylinder(node, nil, nodeIndex)}
	case componentTypeLed:
		return []components.Component{buildLed(node, nil, nodeIndex)}
	case componentTypeMessageQueue, componentTypeQueue:
		return []components.Component{buildMessageQueue(node, nil, nodeIndex)}
	case componentTypeCDN, componentTypeEdge:
		return []components.Component{buildCDN(node, nil, nodeIndex)}
	case componentTypeAPIGateway:
		return []components.Component{buildAPIGateway(node, nil, nodeIndex)}
	case componentTypeBackgroundWorker:
		return []components.Component{buildBackgroundWorker(node, nil, nodeIndex)}
	case componentTypePackage:
		return []components.Component{buildPackage(node, nil, nodeIndex)}
	case componentTypeArtifact, componentTypeFile:
		return []components.Component{buildArtifact(node, nil, nodeIndex)}
	case componentTypeRectangle:
		return []components.Component{buildRectangle(node, nil, nodeIndex)}
	default:
		return []components.Component{buildRectangle(node, nil, nodeIndex)}
	}
}

func buildBrowser(node parser.Node, nodeIndex map[string]components.Shape) components.Component {
	browser := diagrambrowser.New(node.Text)
	browser.Shape = components.Shape{
		Width:  defaultBrowserWidth,
		Height: defaultBrowserHeight,
		X:      defaultComponentX,
		Y:      defaultComponentY,
	}

	applyIDStateProperties(node, &browser.Shape, &browser.Props, node.Text)
	browser.State = applyNamedStateProperties(node, &browser.Shape, &browser.Props, false)

	nodeIndex[node.Text] = browser.Shape
	return browser
}

func buildVM(node parser.Node, nodeIndex map[string]components.Shape) components.Component {
	vm := diagramvm.New(node.Text)
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
	return vm
}

func layoutVMChildren(parent parser.Node, vm *diagramvm.Component, nodeIndex map[string]components.Shape) {
	if len(parent.Children) == 0 {
		return
	}

	for _, child := range parent.Children {
		switch string(child.Type) {
		case componentTypeServer:
			server := buildServer(child, vm, nodeIndex)
			vm.AddChild(server)
		case componentTypeTerminal:
			terminal := buildTerminal(child, vm, nodeIndex)
			vm.AddChild(terminal)
		case componentTypeDatabase:
			database := buildDatabase(child, vm, nodeIndex)
			vm.AddChild(database)
		case componentTypeCylinder:
			cylinder := buildCylinder(child, vm, nodeIndex)
			vm.AddChild(cylinder)
		case componentTypeLed:
			led := buildLed(child, vm, nodeIndex)
			vm.AddChild(led)
		case componentTypeMessageQueue, componentTypeQueue:
			queue := buildMessageQueue(child, vm, nodeIndex)
			vm.AddChild(queue)
		case componentTypeCDN, componentTypeEdge:
			cdn := buildCDN(child, vm, nodeIndex)
			vm.AddChild(cdn)
		case componentTypeAPIGateway:
			gateway := buildAPIGateway(child, vm, nodeIndex)
			vm.AddChild(gateway)
		case componentTypeBackgroundWorker:
			worker := buildBackgroundWorker(child, vm, nodeIndex)
			vm.AddChild(worker)
		case componentTypePackage:
			pkg := buildPackage(child, vm, nodeIndex)
			vm.AddChild(pkg)
		case componentTypeArtifact, componentTypeFile:
			artifact := buildArtifact(child, vm, nodeIndex)
			vm.AddChild(artifact)
		case componentTypeRectangle:
			rect := buildRectangle(child, vm, nodeIndex)
			vm.AddChild(rect)
		default:
			if child.Type == parser.NODE_ELEMENT {
				rect := buildRectangle(child, vm, nodeIndex)
				vm.AddChild(rect)
				continue
			}
			// silently ignore unknown child types
		}
	}
}

func buildServer(node parser.Node, vm *diagramvm.Component, nodeIndex map[string]components.Shape) *diagramserver.Adapter {
	server := diagramserver.NewAdapter(node.Text)
	shape := components.Shape{
		Width:  defaultServerWidth,
		Height: defaultServerHeight,
		X:      defaultComponentX,
		Y:      defaultComponentY,
	}

	applyIDStateProperties(node, &shape, &server.Comp.Props, node.Text)
	server.State = applyNamedStateProperties(node, &shape, &server.Comp.Props, true)

	absServerShape := shape
	if vm != nil {
		contentOffsetX := vm.Shape.Width * diagramvm.VMContentAreaXRatio
		contentOffsetY := vm.Shape.Height * diagramvm.VMContentAreaYRatio
		server.Comp.SetContainer(coreShape(vm.Shape.X+contentOffsetX, vm.Shape.Y+contentOffsetY, vm.Shape.Width*diagramvm.VMContentAreaWidthRatio, vm.Shape.Height*diagramvm.VMContentAreaHeightRatio))
		absServerShape.X = vm.Shape.X + contentOffsetX + absServerShape.X
		absServerShape.Y = vm.Shape.Y + contentOffsetY + absServerShape.Y
	}
	server.SetShape(shape)
	nodeIndex[node.Text] = absServerShape

	return server
}

func buildTerminal(node parser.Node, vm *diagramvm.Component, nodeIndex map[string]components.Shape) *diagramterminal.Adapter {
	terminal := diagramterminal.NewAdapter(node.Text)
	shape := components.Shape{
		Width:  defaultTerminalWidth,
		Height: defaultTerminalHeight,
		X:      defaultComponentX,
		Y:      defaultComponentY,
	}

	applyIDStateProperties(node, &shape, &terminal.Comp.Props, node.Text)
	terminal.State = applyNamedStateProperties(node, &shape, &terminal.Comp.Props, true)

	absShape := shape
	if vm != nil {
		contentOffsetX := vm.Shape.Width * diagramvm.VMContentAreaXRatio
		contentOffsetY := vm.Shape.Height * diagramvm.VMContentAreaYRatio
		terminal.Comp.SetContainer(coreShape(vm.Shape.X+contentOffsetX, vm.Shape.Y+contentOffsetY, vm.Shape.Width*diagramvm.VMContentAreaWidthRatio, vm.Shape.Height*diagramvm.VMContentAreaHeightRatio))
		absShape.X = vm.Shape.X + contentOffsetX + absShape.X
		absShape.Y = vm.Shape.Y + contentOffsetY + absShape.Y
	}
	terminal.SetShape(shape)
	nodeIndex[node.Text] = absShape

	return terminal
}

func buildCylinder(node parser.Node, vm *diagramvm.Component, nodeIndex map[string]components.Shape) *diagramcylinder.Component {
	cylinder := diagramcylinder.New(node.Text)
	cylinder.Shape = components.Shape{
		Width:  defaultCylinderWidth,
		Height: defaultCylinderHeight,
		X:      defaultComponentX,
		Y:      defaultComponentY,
	}

	applyIDStateProperties(node, &cylinder.Shape, &cylinder.Props, node.Text)
	_ = applyNamedStateProperties(node, &cylinder.Shape, &cylinder.Props, true)

	absShape := cylinder.Shape
	if vm != nil {
		contentOffsetX := vm.Shape.Width * diagramvm.VMContentAreaXRatio
		contentOffsetY := vm.Shape.Height * diagramvm.VMContentAreaYRatio
		absShape.X = vm.Shape.X + contentOffsetX + absShape.X
		absShape.Y = vm.Shape.Y + contentOffsetY + absShape.Y
	}
	nodeIndex[node.Text] = absShape

	return cylinder
}

func buildLed(node parser.Node, vm *diagramvm.Component, nodeIndex map[string]components.Shape) *diagramled.Component {
	led := diagramled.New(node.Text)
	led.Shape = components.Shape{
		Width:  defaultLedWidth,
		Height: defaultLedHeight,
		X:      defaultComponentX,
		Y:      defaultComponentY,
	}

	applyIDStateProperties(node, &led.Shape, &led.Props, node.Text)
	_ = applyNamedStateProperties(node, &led.Shape, &led.Props, true)

	absShape := led.Shape
	if vm != nil {
		contentOffsetX := vm.Shape.Width * diagramvm.VMContentAreaXRatio
		contentOffsetY := vm.Shape.Height * diagramvm.VMContentAreaYRatio
		absShape.X = vm.Shape.X + contentOffsetX + absShape.X
		absShape.Y = vm.Shape.Y + contentOffsetY + absShape.Y
	}
	nodeIndex[node.Text] = absShape

	return led
}

func buildDatabase(node parser.Node, vm *diagramvm.Component, nodeIndex map[string]components.Shape) *diagramdatabase.Component {
	database := diagramdatabase.New(node.Text)
	database.Shape = components.Shape{
		Width:  defaultDatabaseWidth,
		Height: defaultDatabaseHeight,
		X:      defaultComponentX,
		Y:      defaultComponentY,
	}

	applyIDStateProperties(node, &database.Shape, &database.Props, node.Text)
	database.State = applyNamedStateProperties(node, &database.Shape, &database.Props, true)

	absShape := database.Shape
	if vm != nil {
		contentOffsetX := vm.Shape.Width * diagramvm.VMContentAreaXRatio
		contentOffsetY := vm.Shape.Height * diagramvm.VMContentAreaYRatio
		absShape.X = vm.Shape.X + contentOffsetX + absShape.X
		absShape.Y = vm.Shape.Y + contentOffsetY + absShape.Y
	}
	nodeIndex[node.Text] = absShape

	return database
}

func buildMessageQueue(node parser.Node, vm *diagramvm.Component, nodeIndex map[string]components.Shape) *diagrammessagequeue.Component {
	queue := diagrammessagequeue.New(node.Text)
	queue.Shape = components.Shape{
		Width:  defaultQueueWidth,
		Height: defaultQueueHeight,
		X:      defaultComponentX,
		Y:      defaultComponentY,
	}

	applyIDStateProperties(node, &queue.Shape, &queue.Props, node.Text)
	queue.State = applyNamedStateProperties(node, &queue.Shape, &queue.Props, true)

	absShape := queue.Shape
	if vm != nil {
		contentOffsetX := vm.Shape.Width * diagramvm.VMContentAreaXRatio
		contentOffsetY := vm.Shape.Height * diagramvm.VMContentAreaYRatio
		absShape.X = vm.Shape.X + contentOffsetX + absShape.X
		absShape.Y = vm.Shape.Y + contentOffsetY + absShape.Y
	}
	nodeIndex[node.Text] = absShape

	return queue
}

func buildCDN(node parser.Node, vm *diagramvm.Component, nodeIndex map[string]components.Shape) *diagramcdn.Component {
	cdn := diagramcdn.New(node.Text)
	cdn.Shape = components.Shape{
		Width:  defaultCDNWidth,
		Height: defaultCDNHeight,
		X:      defaultComponentX,
		Y:      defaultComponentY,
	}

	applyIDStateProperties(node, &cdn.Shape, &cdn.Props, node.Text)
	cdn.State = applyNamedStateProperties(node, &cdn.Shape, &cdn.Props, true)

	absShape := cdn.Shape
	if vm != nil {
		contentOffsetX := vm.Shape.Width * diagramvm.VMContentAreaXRatio
		contentOffsetY := vm.Shape.Height * diagramvm.VMContentAreaYRatio
		absShape.X = vm.Shape.X + contentOffsetX + absShape.X
		absShape.Y = vm.Shape.Y + contentOffsetY + absShape.Y
	}
	nodeIndex[node.Text] = absShape

	return cdn
}

func buildAPIGateway(node parser.Node, vm *diagramvm.Component, nodeIndex map[string]components.Shape) *diagramapigateway.Component {
	gateway := diagramapigateway.New(node.Text)
	gateway.Shape = components.Shape{
		Width:  defaultAPIGatewayWidth,
		Height: defaultAPIGatewayHeight,
		X:      defaultComponentX,
		Y:      defaultComponentY,
	}

	applyIDStateProperties(node, &gateway.Shape, &gateway.Props, node.Text)
	gateway.State = applyNamedStateProperties(node, &gateway.Shape, &gateway.Props, true)

	absShape := gateway.Shape
	if vm != nil {
		contentOffsetX := vm.Shape.Width * diagramvm.VMContentAreaXRatio
		contentOffsetY := vm.Shape.Height * diagramvm.VMContentAreaYRatio
		absShape.X = vm.Shape.X + contentOffsetX + absShape.X
		absShape.Y = vm.Shape.Y + contentOffsetY + absShape.Y
	}
	nodeIndex[node.Text] = absShape

	return gateway
}

func buildBackgroundWorker(node parser.Node, vm *diagramvm.Component, nodeIndex map[string]components.Shape) *diagrambackgroundworker.Component {
	worker := diagrambackgroundworker.New(node.Text)
	worker.Shape = components.Shape{
		Width:  defaultBackgroundWorkerW,
		Height: defaultBackgroundWorkerH,
		X:      defaultComponentX,
		Y:      defaultComponentY,
	}

	applyIDStateProperties(node, &worker.Shape, &worker.Props, node.Text)
	worker.State = applyNamedStateProperties(node, &worker.Shape, &worker.Props, true)

	absShape := worker.Shape
	if vm != nil {
		contentOffsetX := vm.Shape.Width * diagramvm.VMContentAreaXRatio
		contentOffsetY := vm.Shape.Height * diagramvm.VMContentAreaYRatio
		absShape.X = vm.Shape.X + contentOffsetX + absShape.X
		absShape.Y = vm.Shape.Y + contentOffsetY + absShape.Y
	}
	nodeIndex[node.Text] = absShape

	return worker
}

func buildPackage(node parser.Node, vm *diagramvm.Component, nodeIndex map[string]components.Shape) *diagrampackage.Component {
	pkg := diagrampackage.New(node.Text)
	pkg.Shape = components.Shape{
		Width:  defaultPackageWidth,
		Height: defaultPackageHeight,
		X:      defaultComponentX,
		Y:      defaultComponentY,
	}

	applyIDStateProperties(node, &pkg.Shape, &pkg.Props, node.Text)
	pkg.State = applyNamedStateProperties(node, &pkg.Shape, &pkg.Props, true)

	absShape := pkg.Shape
	if vm != nil {
		contentOffsetX := vm.Shape.Width * diagramvm.VMContentAreaXRatio
		contentOffsetY := vm.Shape.Height * diagramvm.VMContentAreaYRatio
		absShape.X = vm.Shape.X + contentOffsetX + absShape.X
		absShape.Y = vm.Shape.Y + contentOffsetY + absShape.Y
	}
	nodeIndex[node.Text] = absShape

	return pkg
}

func buildArtifact(node parser.Node, vm *diagramvm.Component, nodeIndex map[string]components.Shape) *diagramartifact.Component {
	artifact := diagramartifact.New(node.Text)
	artifact.Shape = components.Shape{
		Width:  defaultArtifactWidth,
		Height: defaultArtifactHeight,
		X:      defaultComponentX,
		Y:      defaultComponentY,
	}

	applyIDStateProperties(node, &artifact.Shape, &artifact.Props, node.Text)
	artifact.State = applyNamedStateProperties(node, &artifact.Shape, &artifact.Props, true)

	absShape := artifact.Shape
	if vm != nil {
		contentOffsetX := vm.Shape.Width * diagramvm.VMContentAreaXRatio
		contentOffsetY := vm.Shape.Height * diagramvm.VMContentAreaYRatio
		absShape.X = vm.Shape.X + contentOffsetX + absShape.X
		absShape.Y = vm.Shape.Y + contentOffsetY + absShape.Y
	}
	nodeIndex[node.Text] = absShape

	return artifact
}

func buildRectangle(node parser.Node, vm *diagramvm.Component, nodeIndex map[string]components.Shape) *diagramrectangle.Adapter {
	rect := diagramrectangle.NewAdapter(node.Text)
	shape := components.Shape{
		Width:  defaultServerWidth,
		Height: defaultServerHeight,
		X:      defaultComponentX,
		Y:      defaultComponentY,
	}

	applyIDStateProperties(node, &shape, &rect.Comp.Props, node.Text)
	rect.State = applyNamedStateProperties(node, &shape, &rect.Comp.Props, true)

	absRectShape := shape
	if vm != nil {
		contentOffsetX := vm.Shape.Width * diagramvm.VMContentAreaXRatio
		contentOffsetY := vm.Shape.Height * diagramvm.VMContentAreaYRatio
		rect.Comp.SetContainer(coreShape(vm.Shape.X+contentOffsetX, vm.Shape.Y+contentOffsetY, vm.Shape.Width*diagramvm.VMContentAreaWidthRatio, vm.Shape.Height*diagramvm.VMContentAreaHeightRatio))
		absRectShape.X = vm.Shape.X + contentOffsetX + absRectShape.X
		absRectShape.Y = vm.Shape.Y + contentOffsetY + absRectShape.Y
	}
	rect.SetShape(shape)
	nodeIndex[node.Text] = absRectShape

	return rect
}

func coreShape(x, y, width, height float64) core.Shape {
	return core.Shape{X: x, Y: y, Width: width, Height: height}
}

func applyIDStateProperties(node parser.Node, shape *components.Shape, props propertyParser, componentID string) {
	applyGeometryDefinition(componentID, shape, node.PropsDef)
	parseComponentProps(componentID, props, node.PropsDef)

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
		return
	}
	// Pass empty nodeIndex for now - alignment resolution will happen later
	applyGeometryProps(shape, geometry, make(map[string]components.Shape))
}

func parseComponentProps(target string, parser propertyParser, propsDef string) {
	if parser == nil {
		return
	}
	_ = parser.Parse(propsDef)
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
			continue
		}

		fromAnchor := normalizeAnchor(conn.FromAnchor)
		toAnchor := normalizeAnchor(conn.ToAnchor)
		start := computeAnchorPoint(fromShape, fromAnchor)
		end := computeAnchorPoint(toShape, toAnchor)

		points := routeArrowPoints(start, end, fromAnchor, toAnchor)
		bendPoints := make([]core.Point, 0)
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
		if len(anchor.Directions) == 0 && anchor.Raw != "" {
			anchor.Directions = anchorDirections(anchor)
		}
		return anchor
	}

	normalized := parser.AnchorDescriptor{
		Raw:                   anchor.Raw,
		Directions:            anchorDirections(anchor),
		HorizontalFraction:    anchor.HorizontalFraction,
		VerticalFraction:      anchor.VerticalFraction,
		HasHorizontalFraction: anchor.HasHorizontalFraction,
		HasVerticalFraction:   anchor.HasVerticalFraction,
	}
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

func computeAnchorPoint(shape components.Shape, anchor parser.AnchorDescriptor) core.Point {
	centerX := shape.X + shape.Width*0.5
	centerY := shape.Y + shape.Height*0.5
	point := core.Point{X: centerX, Y: centerY}

	directions := anchorDirections(anchor)
	if len(directions) == 0 {
		switch {
		case anchor.Horizontal < 0:
			point.X = shape.X
			point.Y = verticalEdgeOffsetForAnchor(shape, verticalDirectionFromSign(anchor.Vertical), anchor)
		case anchor.Horizontal > 0:
			point.X = shape.X + shape.Width
			point.Y = verticalEdgeOffsetForAnchor(shape, verticalDirectionFromSign(anchor.Vertical), anchor)
		case anchor.Vertical < 0:
			point.Y = shape.Y
			point.X = horizontalEdgeOffsetForAnchor(shape, horizontalDirectionFromSign(anchor.Horizontal), anchor)
		case anchor.Vertical > 0:
			point.Y = shape.Y + shape.Height
			point.X = horizontalEdgeOffsetForAnchor(shape, horizontalDirectionFromSign(anchor.Horizontal), anchor)
		}
		return point
	}

	primary := directions[0]
	secondary := resolveSecondaryDirection(primary, directions[1:])

	switch {
	case isVerticalDirection(primary):
		point.Y = verticalEdgePosition(shape, primary)
		point.X = horizontalEdgeOffsetForAnchor(shape, secondary, anchor)
	case isHorizontalDirection(primary):
		point.X = horizontalEdgePosition(shape, primary)
		point.Y = verticalEdgeOffsetForAnchor(shape, secondary, anchor)
	default:
		switch {
		case anchor.Horizontal < 0:
			point.X = shape.X
			point.Y = verticalEdgeOffsetForAnchor(shape, verticalDirectionFromSign(anchor.Vertical), anchor)
		case anchor.Horizontal > 0:
			point.X = shape.X + shape.Width
			point.Y = verticalEdgeOffsetForAnchor(shape, verticalDirectionFromSign(anchor.Vertical), anchor)
		case anchor.Vertical < 0:
			point.Y = shape.Y
			point.X = horizontalEdgeOffsetForAnchor(shape, horizontalDirectionFromSign(anchor.Horizontal), anchor)
		case anchor.Vertical > 0:
			point.Y = shape.Y + shape.Height
			point.X = horizontalEdgeOffsetForAnchor(shape, horizontalDirectionFromSign(anchor.Horizontal), anchor)
		}
	}

	return point
}

func routeArrowPoints(start, end core.Point, fromAnchor, toAnchor parser.AnchorDescriptor) []core.Point {
	points := []core.Point{start}

	if floatsNearlyEqual(start.X, end.X) || floatsNearlyEqual(start.Y, end.Y) {
		points = append(points, end)
		return points
	}

	horizontalFirst := shouldRouteHorizontallyFirst(fromAnchor, toAnchor)

	if horizontalFirst {
		direction := resolveAxisDirection(fromAnchor.Horizontal, toAnchor.Horizontal)
		elbowX := start.X + direction*arrowElbowPadding
		points = append(points, core.Point{X: elbowX, Y: start.Y})
		points = append(points, core.Point{X: elbowX, Y: end.Y})
	} else {
		direction := resolveAxisDirection(fromAnchor.Vertical, toAnchor.Vertical)
		elbowY := start.Y + direction*arrowElbowPadding
		points = append(points, core.Point{X: start.X, Y: elbowY})
		points = append(points, core.Point{X: end.X, Y: elbowY})
	}

	points = append(points, end)
	return points
}

func anchorDirections(anchor parser.AnchorDescriptor) []rune {
	if len(anchor.Directions) > 0 {
		return anchor.Directions
	}

	if anchor.Raw != "" {
		lower := strings.ToLower(anchor.Raw)
		directions := make([]rune, 0, len(lower))
		for _, r := range lower {
			switch r {
			case 'n', 's', 'e', 'w':
				directions = append(directions, r)
			}
		}
		if len(directions) > 0 {
			return directions
		}
	}

	directions := make([]rune, 0, 2)
	if anchor.Horizontal < 0 {
		directions = append(directions, 'w')
	} else if anchor.Horizontal > 0 {
		directions = append(directions, 'e')
	}
	if anchor.Vertical < 0 {
		directions = append(directions, 'n')
	} else if anchor.Vertical > 0 {
		directions = append(directions, 's')
	}
	return directions
}

func resolveSecondaryDirection(primary rune, remaining []rune) rune {
	for _, r := range remaining {
		if isVerticalDirection(primary) && isHorizontalDirection(r) {
			return r
		}
		if isHorizontalDirection(primary) && isVerticalDirection(r) {
			return r
		}
	}
	return 0
}

func isVerticalDirection(r rune) bool {
	return r == 'n' || r == 's'
}

func isHorizontalDirection(r rune) bool {
	return r == 'e' || r == 'w'
}

func verticalEdgePosition(shape components.Shape, direction rune) float64 {
	if direction == 'n' {
		return shape.Y
	}
	return shape.Y + shape.Height
}

func horizontalEdgePosition(shape components.Shape, direction rune) float64 {
	if direction == 'w' {
		return shape.X
	}
	return shape.X + shape.Width
}

func horizontalEdgeOffsetForAnchor(shape components.Shape, direction rune, anchor parser.AnchorDescriptor) float64 {
	if anchor.HasHorizontalFraction {
		return shape.X + shape.Width*anchor.HorizontalFraction
	}

	switch direction {
	case 'w':
		return shape.X + shape.Width*0.25
	case 'e':
		return shape.X + shape.Width*0.75
	default:
		return shape.X + shape.Width*0.5
	}
}

func verticalEdgeOffsetForAnchor(shape components.Shape, direction rune, anchor parser.AnchorDescriptor) float64 {
	if anchor.HasVerticalFraction {
		return shape.Y + shape.Height*anchor.VerticalFraction
	}

	switch direction {
	case 'n':
		return shape.Y + shape.Height*0.25
	case 's':
		return shape.Y + shape.Height*0.75
	default:
		return shape.Y + shape.Height*0.5
	}
}

func horizontalDirectionFromSign(value float64) rune {
	switch {
	case value < 0:
		return 'w'
	case value > 0:
		return 'e'
	default:
		return 0
	}
}

func verticalDirectionFromSign(value float64) rune {
	switch {
	case value < 0:
		return 'n'
	case value > 0:
		return 's'
	default:
		return 0
	}
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
