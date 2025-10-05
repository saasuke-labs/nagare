package diagram

import (
	_ "embed"
	"testing"
)

//go:embed fixtures/code_block_1.txt
var codeBlock1 string

//go:embed fixtures/svg_1.svg
var svg1 string

//go:embed fixtures/code_block_anchor_percent.txt
var codeBlockAnchorPercent string

//go:embed fixtures/svg_anchor_percent.svg
var svgAnchorPercent string

func TestCreateDiagramFromActualCodeBlocks(t *testing.T) {
	testData := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "code block 1",
			code:     codeBlock1,
			expected: svg1,
		},
		{
			name:     "fractional anchors",
			code:     codeBlockAnchorPercent,
			expected: svgAnchorPercent,
		},
	}

	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			html, err := CreateDiagram(td.code)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if html != td.expected {
				t.Fatalf("expected HTML does not match actual.\nExpected:\n%s\n\nGot:\n%s", td.expected, html)
			}
		})
	}

}
