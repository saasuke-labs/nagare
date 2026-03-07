package browser

import (
	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/core"
)

type Legacy struct {
	*components.Browser
}

func NewLegacy(id string) *Legacy {
	b := components.NewBrowser()
	b.Text = id
	return &Legacy{Browser: b}
}

func (l *Legacy) SetShape(shape components.Shape) {
	l.Shape = shape
}

const (
	DefaultWidth  = 640.0
	DefaultHeight = 420.0
)

func DrawFromRenderNode(id string, nodeProps map[string]any) string {
	comp := NewLegacy(id)
	comp.Shape = core.ShapeFromProps(nodeProps, DefaultWidth, DefaultHeight)
	if raw, ok := nodeProps["_rawProps"].(string); ok {
		_ = comp.Props.Parse(raw)
	}
	return comp.Draw()
}
