package core

import "strings"

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

// RequestResponse is the connector every component gets by default: a
// "request" action draws a solid arrow to its target and a "response" action a
// dashed one. Other action names produce nothing.
//
//	@api.request(target: @db)            // ports inferred from positions
//	@api.response(target: @shop, dir: rl) // ports from a direction hint
type RequestResponse struct{}

// ResolveAction implements ActionConnector.
func (RequestResponse) ResolveAction(sourceID, actionName string, actionProps map[string]any) []ConnectionIntent {
	if actionName != "request" && actionName != "response" {
		return nil
	}
	targetRaw, _ := actionProps["target"].(string)
	targetID := strings.TrimPrefix(strings.TrimSpace(targetRaw), "@")
	if targetID == "" {
		return nil
	}
	dir, _ := actionProps["dir"].(string)
	style := ""
	if actionName == "response" {
		style = "dashed"
	}
	return []ConnectionIntent{{
		FromID: sourceID,
		ToID:   targetID,
		Dir:    dir,
		Style:  style,
	}}
}
