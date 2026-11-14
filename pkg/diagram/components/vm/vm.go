package vm

import (
	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/core"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/style"
)

// The content-area ratios are part of the layout contract: children of a VM
// are positioned with them, so the chrome must leave exactly this area free.
const (
	VMContentAreaXRatio      = components.VMContentAreaXRatio
	VMContentAreaYRatio      = components.VMContentAreaYRatio
	VMContentAreaWidthRatio  = components.VMContentAreaWidthRatio
	VMContentAreaHeightRatio = components.VMContentAreaHeightRatio
	DefaultWidth             = 640.0
	DefaultHeight            = 420.0
)

type Component struct {
	*components.VM
	Theme style.Theme
}

func New(id string) *Component {
	v := components.NewVM()
	v.Text = id
	// Colours are theme-driven; props only override when set explicitly.
	v.Props = components.VMProps{}
	return &Component{VM: v, Theme: style.Light}
}

func (l *Component) SetShape(shape components.Shape) {
	l.Shape = shape
}

func (l *Component) Draw() string {
	w, h := l.Width, l.Height
	return style.Window(style.WindowSpec{
		X: l.X, Y: l.Y, W: w, H: h,
		Class: "nagare-vm",
		BarH:  h * VMContentAreaYRatio,
		CX:    w * VMContentAreaXRatio, CY: h * VMContentAreaYRatio,
		CW: w * VMContentAreaWidthRatio, CH: h * VMContentAreaHeightRatio,
		Title:       style.Or(l.Props.Title, l.Text),
		Bg:          l.Props.BackgroundColor,
		Fg:          l.Props.ForegroundColor,
		ContentFill: style.Or(l.Props.ContentBackgroundColor, l.Theme.Inset),
		Theme:       l.Theme,
	}) + style.WindowEnd
}

func DrawFromRenderNode(id string, nodeProps map[string]any) string {
	comp := New(id)
	comp.Shape = core.ShapeFromProps(nodeProps, DefaultWidth, DefaultHeight)
	comp.Theme = style.ThemeFromProps(nodeProps)
	if raw, ok := nodeProps["_rawProps"].(string); ok {
		_ = comp.Props.Parse(raw)
	}
	return comp.Draw()
}
