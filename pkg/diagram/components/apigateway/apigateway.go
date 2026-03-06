package apigateway

import "github.com/saasuke-labs/nagare/pkg/components"

type Legacy struct {
	*components.APIGateway
}

func NewLegacy(id string) *Legacy {
	return &Legacy{APIGateway: components.NewAPIGateway(id)}
}

func (l *Legacy) SetShape(shape components.Shape) { l.Shape = shape }
func (l *Legacy) Offset(dx, dy float64) {
	l.X -= dx
	l.Y -= dy
}
