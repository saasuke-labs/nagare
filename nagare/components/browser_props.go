package components

import "nagare/props"

// BrowserProps defines the configurable properties for a Browser component
type BrowserProps struct {
	URL             string `prop:"url"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	Text            string `prop:"text"`
}

// Parse implements the Props interface
func (b *BrowserProps) Parse(input string) error {
	return props.ParseProps(input, b)
}

// DefaultBrowserProps returns a BrowserProps with default values
func DefaultBrowserProps() BrowserProps {
	return BrowserProps{
		URL:             "",
		BackgroundColor: "#e6f3ff",
		ForegroundColor: "#333333",
		Text:            "",
	}
}
