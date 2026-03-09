package browser

import (
	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/core"
)

type Component struct {
	*components.Browser
}

func New(id string) *Component {
	b := components.NewBrowser()
	b.Text = id
	return &Component{Browser: b}
}

func (l *Component) SetShape(shape components.Shape) {
	l.Shape = shape
}

const (
	DefaultWidth  = 640.0
	DefaultHeight = 420.0
)

func DrawFromRenderNode(id string, nodeProps map[string]any) string {
	comp := New(id)
	comp.Shape = core.ShapeFromProps(nodeProps, DefaultWidth, DefaultHeight)
	if raw, ok := nodeProps["_rawProps"].(string); ok {
		_ = comp.Props.Parse(raw)
	}
	return comp.Draw()
}
