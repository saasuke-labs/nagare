package vm

import "github.com/saasuke-labs/nagare/pkg/components"

const (
	VMContentAreaXRatio      = components.VMContentAreaXRatio
	VMContentAreaYRatio      = components.VMContentAreaYRatio
	VMContentAreaWidthRatio  = components.VMContentAreaWidthRatio
	VMContentAreaHeightRatio = components.VMContentAreaHeightRatio
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
