package layout

import (
	"fmt"
	"nagare/components"
	"nagare/parser"
	"nagare/props"
	"strings"
)

const (
	defaultBrowserWidth  = 640.0
	defaultBrowserHeight = 420.0
	defaultVMWidth       = 640.0
	defaultVMHeight      = 420.0
	defaultServerWidth   = 200.0
	defaultServerHeight  = 140.0
)

// Rect represents a rectangle in the layout
type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// Layout represents the computed layout of an element
type Layout struct {
	Bounds   Rect
	Children []components.Component
}

type geometryProps struct {
	X      *int `prop:"x"`
	Y      *int `prop:"y"`
	Width  *int `prop:"w"`
	Height *int `prop:"h"`
}

func parseGeometry(def string) (geometryProps, error) {
	geom := geometryProps{}
	if strings.TrimSpace(def) == "" {
		return geom, nil
	}
	if err := props.ParseProps(def, &geom); err != nil {
		return geom, err
	}
	return geom, nil
}

func applyGeometry(shape *components.Shape, geom geometryProps) {
	if geom.Width != nil {
		shape.Width = float64(*geom.Width)
	}
	if geom.Height != nil {
		shape.Height = float64(*geom.Height)
	}
	if geom.X != nil {
		shape.X = float64(*geom.X)
	}
	if geom.Y != nil {
		shape.Y = float64(*geom.Y)
	}
}

// Calculate computes the layout for an AST
func Calculate(node parser.Node, canvasWidth, canvasHeight float64) Layout {
	boundsWidth := canvasWidth
	boundsHeight := canvasHeight

	if layoutState, ok := node.Globals["layout"]; ok {
		if geom, err := parseGeometry(layoutState.PropsDef); err == nil {
			if geom.Width != nil {
				boundsWidth = float64(*geom.Width)
			}
			if geom.Height != nil {
				boundsHeight = float64(*geom.Height)
			}
		} else {
			fmt.Printf("failed to parse @layout props: %v\n", err)
		}
	}

	children := make([]components.Component, 0, len(node.Children))

	for _, child := range node.Children {
		switch child.Type {
		case "Browser":
			browser := components.NewBrowser()
			browser.Shape = components.Shape{
				Width:  defaultBrowserWidth,
				Height: defaultBrowserHeight,
				X:      0,
				Y:      0,
			}

			if idState, ok := child.States[child.Text]; ok {
				if geom, err := parseGeometry(idState.PropsDef); err == nil {
					applyGeometry(&browser.Shape, geom)
				} else {
					fmt.Printf("failed to parse geometry for %s: %v\n", child.Text, err)
				}

				if err := browser.Props.Parse(idState.PropsDef); err != nil {
					fmt.Printf("failed to parse props for %s: %v\n", child.Text, err)
				}
			}

			if child.State != "" {
				if state, ok := child.States[child.State]; ok {
					browser.State = state.Name
					if err := browser.Props.Parse(state.PropsDef); err != nil {
						fmt.Printf("failed to parse props for state %s: %v\n", state.Name, err)
					}
				}
			}

			children = append(children, browser)
			fmt.Printf("State: %s, Props: %+v\n", browser.State, browser.Props)
		case "VM":
			vm := components.NewVM()
			vm.Shape = components.Shape{
				Width:  defaultVMWidth,
				Height: defaultVMHeight,
				X:      0,
				Y:      0,
			}

			if idState, ok := child.States[child.Text]; ok {
				if geom, err := parseGeometry(idState.PropsDef); err == nil {
					applyGeometry(&vm.Shape, geom)
				} else {
					fmt.Printf("failed to parse geometry for %s: %v\n", child.Text, err)
				}

				if err := vm.Props.Parse(idState.PropsDef); err != nil {
					fmt.Printf("failed to parse props for %s: %v\n", child.Text, err)
				}
			}

			if child.State != "" {
				if state, ok := child.States[child.State]; ok {
					vm.State = state.Name
					if err := vm.Props.Parse(state.PropsDef); err != nil {
						fmt.Printf("failed to parse props for state %s: %v\n", state.Name, err)
					}
				}
			}

			if len(child.Children) > 0 {
				childComponents := make([]components.Component, 0, len(child.Children))

				for _, grandchild := range child.Children {
					switch grandchild.Type {
					case "Server":
						server := components.NewServer(grandchild.Text)
						server.Shape = components.Shape{
							Width:  defaultServerWidth,
							Height: defaultServerHeight,
							X:      0,
							Y:      0,
						}

						if idState, ok := grandchild.States[grandchild.Text]; ok {
							if geom, err := parseGeometry(idState.PropsDef); err == nil {
								applyGeometry(&server.Shape, geom)
							} else {
								fmt.Printf("failed to parse geometry for %s: %v\n", grandchild.Text, err)
							}

							if err := server.Props.Parse(idState.PropsDef); err != nil {
								fmt.Printf("failed to parse props for %s: %v\n", grandchild.Text, err)
							}
						}

						if grandchild.State != "" {
							if state, ok := grandchild.States[grandchild.State]; ok {
								server.State = state.Name
								if geom, err := parseGeometry(state.PropsDef); err == nil {
									applyGeometry(&server.Shape, geom)
								} else {
									fmt.Printf("failed to parse geometry for state %s: %v\n", state.Name, err)
								}
								if err := server.Props.Parse(state.PropsDef); err != nil {
									fmt.Printf("failed to parse props for state %s: %v\n", state.Name, err)
								}
							}
						}

						childComponents = append(childComponents, server)
					default:
						fmt.Printf("Unknown child type: %s\n", grandchild.Type)
					}
				}
				vm.Children = childComponents
			}

			children = append(children, vm)
			fmt.Printf("State: %s, Props: %+v\n", vm.State, vm.Props)
		default:
			children = append(children, &components.Rectangle{
				Shape: components.Shape{
					Width:  defaultServerWidth,
					Height: defaultServerHeight,
					X:      0,
					Y:      0,
				},
				Text: child.Text,
			})
		}
	}

	return Layout{
		Bounds: Rect{
			X:      0,
			Y:      0,
			Width:  boundsWidth,
			Height: boundsHeight,
		},
		Children: children,
	}
}
