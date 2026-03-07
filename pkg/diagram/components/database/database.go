package database

import (
	"fmt"
	"math"

	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/core"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/cylinder"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/led"
	"github.com/saasuke-labs/nagare/pkg/props"
)

const AllowedActions = "read,write"

const (
	DefaultWidth  = 200.0
	DefaultHeight = 200.0
)

type Props struct {
	Title           string `prop:"title"`
	Engine          string `prop:"engine"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	AccentColor     string `prop:"accent"`
}

func (p *Props) Parse(input string) error {
	return props.ParseProps(input, p)
}

func DefaultProps() Props {
	return Props{
		Title:           "Database",
		Engine:          "PostgreSQL",
		BackgroundColor: "#0f766e",
		ForegroundColor: "#ecfdf5",
		AccentColor:     "#14b8a6",
	}
}

type Legacy struct {
	components.Shape
	Text    string
	Props   Props
	State   string
	Actions map[string][]map[string]any
}

func NewLegacy(id string) *Legacy {
	return &Legacy{Text: id, Props: DefaultProps(), Actions: make(map[string][]map[string]any)}
}

func (l *Legacy) Draw() string {
	return drawComposite(l.Text, l.Shape, l.Props, l.Actions)
}

func (l *Legacy) SetShape(shape components.Shape) { l.Shape = shape }

func (l *Legacy) Offset(dx, dy float64) {
	l.X -= dx
	l.Y -= dy
}

func DrawFromRenderNode(id string, nodeProps map[string]any) string {
	comp := NewLegacy(id)
	comp.Shape = core.ShapeFromProps(nodeProps, DefaultWidth, DefaultHeight)
	_ = comp.Props.Parse(propsRaw(nodeProps))
	comp.Actions = actionMapFromAny(nodeProps["actions"])
	return comp.Draw()
}

func drawComposite(id string, shape components.Shape, p Props, actions map[string][]map[string]any) string {
	cylinderPropsRaw := fmt.Sprintf("title:%q,subtitle:%q,bg:%q,fg:%q,accent:%q", p.Title, p.Engine, p.BackgroundColor, p.ForegroundColor, p.AccentColor)
	cylinderSVG := cylinder.DrawFromRenderNode(id+"-body", map[string]any{
		"x":         shape.X,
		"y":         shape.Y,
		"w":         shape.Width,
		"h":         shape.Height,
		"_rawProps": cylinderPropsRaw,
	})

	writeX, readX, ledY, ledSize := ledLayout(shape)

	writeLedSVG := led.DrawFromRenderNode(id+"-write", map[string]any{
		"x":         writeX,
		"y":         ledY,
		"w":         ledSize,
		"h":         ledSize,
		"_rawProps": `mode:"red"`,
		"actions": map[string][]map[string]any{
			"blink": cloneActionList(actions["write"]),
		},
	})

	readLedSVG := led.DrawFromRenderNode(id+"-read", map[string]any{
		"x":         readX,
		"y":         ledY,
		"w":         ledSize,
		"h":         ledSize,
		"_rawProps": `mode:"green"`,
		"actions": map[string][]map[string]any{
			"blink": cloneActionList(actions["read"]),
		},
	})

	return cylinderSVG + writeLedSVG + readLedSVG
}

func ledLayout(shape components.Shape) (writeX, readX, y, size float64) {
	size = math.Max(8, math.Min(shape.Width, shape.Height)*0.10)
	gap := size * 0.45
	rightPadding := size * 0.9
	y = shape.Y + math.Max(8, shape.Height*0.10)

	readX = shape.X + shape.Width - rightPadding - size
	writeX = readX - gap - size

	minX := shape.X + size*0.4
	if writeX < minX {
		writeX = minX
		readX = writeX + size + gap
	}

	return writeX, readX, y, size
}

func cloneActionList(items []map[string]any) []map[string]any {
	if len(items) == 0 {
		return nil
	}
	copied := make([]map[string]any, 0, len(items))
	for _, item := range items {
		dup := make(map[string]any, len(item))
		for key, val := range item {
			dup[key] = val
		}
		copied = append(copied, dup)
	}
	return copied
}

func BuildLegacy(id string, parent *components.Shape) *Legacy {
	comp := NewLegacy(id)
	comp.Shape = components.Shape{Width: DefaultWidth, Height: DefaultHeight}
	if parent != nil {
		comp.Shape.X = parent.X + comp.Shape.X
		comp.Shape.Y = parent.Y + comp.Shape.Y
	}
	return comp
}

func propsRaw(nodeProps map[string]any) string {
	if v, ok := nodeProps["_rawProps"].(string); ok {
		return v
	}
	return ""
}

func actionMapFromAny(v any) map[string][]map[string]any {
	out := make(map[string][]map[string]any)
	typed, ok := v.(map[string][]map[string]any)
	if !ok {
		return out
	}
	for k, items := range typed {
		copied := make([]map[string]any, 0, len(items))
		for _, item := range items {
			dup := make(map[string]any, len(item))
			for key, val := range item {
				dup[key] = val
			}
			copied = append(copied, dup)
		}
		out[k] = copied
	}
	return out
}
