package messagequeue

import "github.com/saasuke-labs/nagare/pkg/components"

type Legacy struct {
	*components.MessageQueue
}

func NewLegacy(id string) *Legacy {
	return &Legacy{MessageQueue: components.NewMessageQueue(id)}
}

func (l *Legacy) SetShape(shape components.Shape) { l.Shape = shape }
func (l *Legacy) Offset(dx, dy float64) {
	l.X -= dx
	l.Y -= dy
}
