package database

import "github.com/saasuke-labs/nagare/pkg/components"

type Legacy struct {
	*components.Database
}

func NewLegacy(id string) *Legacy {
	return &Legacy{Database: components.NewDatabase(id)}
}

func (l *Legacy) SetShape(shape components.Shape) { l.Shape = shape }
func (l *Legacy) Offset(dx, dy float64) {
	l.X -= dx
	l.Y -= dy
}
