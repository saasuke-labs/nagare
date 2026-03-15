package core

// ConnectionIntent describes an arrow that a component action wants to create.
// Layout resolves coordinates and creates the concrete Arrow geometry so that
// the component never needs to compute canvas-space positions.
type ConnectionIntent struct {
	FromID   string // source component ID
	ToID     string // destination component ID
	FromPort string // optional compass anchor ("e","w","n","s"); inferred when blank
	ToPort   string // optional compass anchor; inferred when blank
	Dir      string // optional direction hint ("lr","rl","tb","bt"); used when ports are blank
	Style    string // "" = solid line, "dashed" = dashed line
}

// ActionConnector is implemented by diagram components whose actions can
// generate arrows. Layout calls ResolveAction for each action-triggered global
// state; the component maps the action to zero or more ConnectionIntents that
// layout resolves into concrete Arrow geometry.
//
// The component is responsible only for the semantic mapping (e.g.
// "request" → solid line, "response" → dashed line). Layout remains
// agnostic to those semantics.
type ActionConnector interface {
	ResolveAction(sourceID, actionName string, actionProps map[string]any) []ConnectionIntent
}
