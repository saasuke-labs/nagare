package components

type HeaderControlProps struct {
	ControlsX      float64
	ControlsY      float64
	ControlRadius  float64
	ControlSpacing float64
}

func NewHeaderControlProps(windowHeight, windowWidth float64) HeaderControlProps {

	controlsX := windowHeight * 0.021875      // 14/640
	controlsY := windowWidth * 0.0333333      // 14/420
	controlRadius := windowHeight * 0.009375  // 6/640
	controlSpacing := windowHeight * 0.028125 // 18/640

	return HeaderControlProps{
		ControlsX:      controlsX,
		ControlsY:      controlsY,
		ControlRadius:  controlRadius,
		ControlSpacing: controlSpacing,
	}
}
