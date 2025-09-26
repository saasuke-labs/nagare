package components

import "fmt"

type Browser struct {
	Shape
	Text string
}

func (r *Browser) Draw(colWidth, rowHeight float64) string {
	actualWidth := float64(r.Width) * colWidth
	actualHeight := float64(r.Height) * rowHeight

	fmt.Println("Drawing browser:", r.Text, "at", r.X, r.Y, "size", r.Width, r.Height)
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

	// Text positioning and size
	textY := actualHeight * 0.0523809 // 22/420
	fontSize := actualWidth * 0.05    // 13/640

	return fmt.Sprintf(
		`<g transform="translate(%f,%f)">
		<g class="ns" filter="url(#softShadow)">
			<rect x="0" y="0" width="%f" height="%f" rx="%f" fill="#f8f9fb" stroke="#333333"/>
			<rect x="0" y="0" width="%f" height="%f" rx="%f" ry="%f" fill="#e9eef7" stroke="#333333"/>
			<rect x="%f" y="%f" width="%f" height="%f" rx="%f" fill="#fff" stroke="#aaaaaa" opacity="0.85"/>
			<rect x="%f" y="%f" width="%f" height="%f" rx="%f" fill="#fff" stroke="#aaaaaa" opacity="0.9"/>
		</g>
		<text x="%f" y="%f" text-anchor="middle" dominant-baseline="middle"
				font-family="-apple-system, Segoe UI, Roboto, Helvetica, Arial, sans-serif"
				font-size="%f" fill="#111">https://app.nagare.local</text>
		<g transform="translate(%f,%f)">
			<circle r="%f" cx="0"  cy="%f" fill="#ff5f57"/>
			<circle r="%f" cx="%f" cy="%f" fill="#febc2e"/>
			<circle r="%f" cx="%f" cy="%f" fill="#28c840"/>
		</g>
	</g>`,
		float64(r.X)*colWidth, float64(r.Y)*rowHeight,
		actualWidth, actualHeight, cornerRadius,
		actualWidth, topBarHeight, cornerRadius, cornerRadius,
		urlBarX, urlBarY, urlBarWidth, urlBarHeight, cornerRadius*0.6,
		contentAreaX, contentAreaY, contentAreaWidth, contentAreaHeight, cornerRadius*0.6,
		actualWidth/2, textY, fontSize,
		controlsX, controlsY,
		controlRadius, controlRadius*1.33,
		controlRadius, controlSpacing, controlRadius*1.33,
		controlRadius, controlSpacing*2, controlRadius*1.33,
	)

}
