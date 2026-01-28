//go:build !webp
// +build !webp

package diagram

import (
	"fmt"
)

// CreateDiagramWebP generates a diagram identical to CreateDiagram but returns it encoded as a WebP image.
// This stub version is used when the webp build tag is not enabled.
func CreateDiagramWebP(code string) ([]byte, error) {
	return nil, fmt.Errorf("webp support is not enabled; rebuild with -tags=webp")
}
