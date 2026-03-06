package browser

import "github.com/saasuke-labs/nagare/pkg/components"

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
