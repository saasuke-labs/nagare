package vm

import (
	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/core"
)

const (
	VMContentAreaXRatio      = components.VMContentAreaXRatio
	VMContentAreaYRatio      = components.VMContentAreaYRatio
	VMContentAreaWidthRatio  = components.VMContentAreaWidthRatio
	VMContentAreaHeightRatio = components.VMContentAreaHeightRatio
	DefaultWidth             = 640.0
	DefaultHeight            = 420.0
)

type Legacy struct {
	*components.VM
}

func NewLegacy(id string) *Legacy {
	v := components.NewVM()
	v.Text = id
	return &Legacy{VM: v}
}

func (l *Legacy) SetShape(shape components.Shape) {
	l.Shape = shape
}

func DrawFromRenderNode(id string, nodeProps map[string]any) string {
	comp := NewLegacy(id)
	comp.Shape = core.ShapeFromProps(nodeProps, DefaultWidth, DefaultHeight)
	if raw, ok := nodeProps["_rawProps"].(string); ok {
		_ = comp.Props.Parse(raw)
	}
	return comp.Draw()
}
