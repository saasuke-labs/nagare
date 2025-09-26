package components

import "fmt"

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

func (r *Browser) Draw(colWidth, rowHeight float64) string {
	actualWidth := float64(r.Width) * colWidth
	actualHeight := float64(r.Height) * rowHeight

	fmt.Println("Drawing browser at", r.X, r.Y, "size", r.Width, r.Height)

	// Calculate proportional dimensions
	cornerRadius := actualWidth * 0.015625   // 10/640
	topBarHeight := actualHeight * 0.1047619 // 44/420
	urlBarHeight := actualHeight * 0.0571428 // 24/420
	urlBarWidth := actualWidth * 0.75        // 480/640
	urlBarX := actualWidth * 0.15625         // 100/640
	urlBarY := actualHeight * 0.0238095      // 10/420

	contentAreaWidth := actualWidth * 0.9625      // 616/640
	contentAreaHeight := actualHeight * 0.8380952 // 352/420
	contentAreaX := actualWidth * 0.01875         // 12/640
	contentAreaY := actualHeight * 0.1333333      // 56/420

	// Window controls
	controlsX := actualWidth * 0.021875      // 14/640
	controlsY := actualHeight * 0.0333333    // 14/420
	controlRadius := actualWidth * 0.009375  // 6/640
	controlSpacing := actualWidth * 0.028125 // 18/640

	// Text size
	fontSize := actualWidth * 0.05 // 13/640

	return fmt.Sprintf(`<g transform="translate(%f,%f)">
                <g class="ns" filter="url(#softShadow)">
                        <rect x="0" y="0" width="%f" height="%f" rx="%f" fill="%s" stroke="%s"/>
                        <rect x="0" y="0" width="%f" height="%f" rx="%f" ry="%f" fill="%s" stroke="%s"/>
                        <rect x="%f" y="%f" width="%f" height="%f" rx="%f" fill="#fff" stroke="%s" opacity="0.85"/>
                        <rect x="%f" y="%f" width="%f" height="%f" rx="%f" fill="%s" stroke="%s" opacity="0.9"/>
                </g>
                <text x="%f" y="%f" text-anchor="middle" dominant-baseline="middle"
                        font-family="-apple-system, Segoe UI, Roboto, Helvetica, Arial, sans-serif"
                        font-size="%f" fill="%s">%s</text>
                <text x="%f" y="%f" text-anchor="middle" dominant-baseline="middle"
                        font-family="-apple-system, Segoe UI, Roboto, Helvetica, Arial, sans-serif"
                        font-size="%f" fill="%s">%s</text>
                <g transform="translate(%f,%f)">
                        <circle r="%f" cx="0"  cy="%f" fill="#ff5f57"/>
                        <circle r="%f" cx="%f" cy="%f" fill="#febc2e"/>
                        <circle r="%f" cx="%f" cy="%f" fill="#28c840"/>
                </g>
        </g>`,
		float64(r.X)*colWidth, float64(r.Y)*rowHeight,
		actualWidth, actualHeight, cornerRadius, r.Props.BackgroundColor, r.Props.ForegroundColor,
		actualWidth, topBarHeight, cornerRadius, cornerRadius, r.Props.BackgroundColor, r.Props.ForegroundColor,
		urlBarX, urlBarY, urlBarWidth, urlBarHeight, cornerRadius*0.6, r.Props.ForegroundColor,
		contentAreaX, contentAreaY, contentAreaWidth, contentAreaHeight, cornerRadius*0.6, r.Props.BackgroundColor, r.Props.ForegroundColor,
		urlBarX+urlBarWidth/2, urlBarY+urlBarHeight/2, fontSize*0.8, r.Props.ForegroundColor, r.Props.URL,
		actualWidth/2, contentAreaY+contentAreaHeight/2, fontSize, r.Props.ForegroundColor, r.Props.Text,
		controlsX, controlsY,
		controlRadius, controlRadius*1.33,
		controlRadius, controlSpacing, controlRadius*1.33,
		controlRadius, controlSpacing*2, controlRadius*1.33,
	)
}
