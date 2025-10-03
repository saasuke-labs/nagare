package components

import (
	"fmt"

	"github.com/saasuke-labs/nagare/pkg/props"
)

// BrowserProps defines the configurable properties for a Browser component
type BrowserProps struct {
	URL                    string `prop:"url"`
	BackgroundColor        string `prop:"bg"`
	ForegroundColor        string `prop:"fg"`
	ContentBackgroundColor string `prop:"contentBg"`
	Text                   string `prop:"text"`
}

// Parse implements the Props interface
func (b *BrowserProps) Parse(input string) error {
	return props.ParseProps(input, b)
}

// DefaultBrowserProps returns a BrowserProps with default values
func DefaultBrowserProps() BrowserProps {
	return BrowserProps{
		URL:                    "",
		BackgroundColor:        "#e6f3ff",
		ForegroundColor:        "#333333",
		ContentBackgroundColor: "#ffffff", // White content area by default
		Text:                   "",
	}
}

type Browser struct {
	Shape
	Text  string
	Props BrowserProps
	State string // Current state name
}

// NewBrowser creates a new Browser with default props
func NewBrowser() *Browser {
	return &Browser{
		Props: DefaultBrowserProps(),
	}
}

type BrowserTemplateData struct {
	X                      float64
	Y                      float64
	Width                  float64
	Height                 float64
	CornerRadius           float64
	TopBarHeight           float64
	UrlBarWidth            float64
	UrlBarHeight           float64
	UrlBarX                float64
	UrlBarY                float64
	ContentAreaWidth       float64
	ContentAreaHeight      float64
	ContentAreaX           float64
	ContentAreaY           float64
	FontSize               float64
	BackgroundColor        string
	ForegroundColor        string
	ContentBackgroundColor string
	URL                    string
	Text                   string
	HeaderControlProps     HeaderControlProps
}

func (r *Browser) Draw() string {
	fmt.Println("Drawing browser at", r.X, r.Y, "size", r.Width, r.Height)

	actualWidth := r.Width
	actualHeight := r.Height

	// Calculate all dimensions
	cornerRadius := actualWidth * 0.015625        // 10/640
	topBarHeight := actualHeight * 0.1047619      // 44/420
	urlBarHeight := actualHeight * 0.0571428      // 24/420
	urlBarWidth := actualWidth * 0.75             // 480/640
	urlBarX := actualWidth * 0.15625              // 100/640
	urlBarY := actualHeight * 0.0238095           // 10/420
	contentAreaWidth := actualWidth * 0.9625      // 616/640
	contentAreaHeight := actualHeight * 0.8380952 // 352/420
	contentAreaX := actualWidth * 0.01875         // 12/640
	contentAreaY := actualHeight * 0.1333333      // 56/420
	fontSize := actualWidth * 0.05                // 13/640

	headerControlProps := NewHeaderControlProps(actualWidth, actualHeight)

	// Create template data
	data := BrowserTemplateData{
		X:                      r.X,
		Y:                      r.Y,
		Width:                  actualWidth,
		Height:                 actualHeight,
		CornerRadius:           cornerRadius,
		TopBarHeight:           topBarHeight,
		UrlBarWidth:            urlBarWidth,
		UrlBarHeight:           urlBarHeight,
		UrlBarX:                urlBarX,
		UrlBarY:                urlBarY,
		ContentAreaWidth:       contentAreaWidth,
		ContentAreaHeight:      contentAreaHeight,
		ContentAreaX:           contentAreaX,
		ContentAreaY:           contentAreaY,
		FontSize:               fontSize,
		BackgroundColor:        r.Props.BackgroundColor,
		ForegroundColor:        r.Props.ForegroundColor,
		ContentBackgroundColor: r.Props.ContentBackgroundColor,
		URL:                    r.Props.URL,
		Text:                   r.Props.Text,
		HeaderControlProps:     headerControlProps,
	}

	result, err := RenderTemplate("browser", data)

	if err != nil {
		fmt.Printf("Error rendering template: %v\n", err)
		return ""
	}

	return result
}
