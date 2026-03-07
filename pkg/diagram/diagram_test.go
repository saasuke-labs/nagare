package diagram

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/saasuke-labs/nagare/pkg/components"
)

//go:embed fixtures/code_block_1.txt
var codeBlock1 string

//go:embed fixtures/code_block_anchor_percent.txt
var codeBlockAnchorPercent string

func TestCreateDiagramFromActualCodeBlocks(t *testing.T) {
	testData := []struct {
		name              string
		code              string
		expectedFragments []string
	}{
		{
			name: "code block 1",
			code: codeBlock1,
			expectedFragments: []string{
				"<svg",
				"Blue",
				"ubuntu@multipass",
				"<polyline",
			},
		},
		{
			name: "fractional anchors",
			code: codeBlockAnchorPercent,
			expectedFragments: []string{
				"<svg",
				"Top",
				"Bottom",
				"<polyline",
			},
		},
	}

	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			// Reset arrow counter to ensure consistent IDs
			components.ResetArrowMarkerCounter()

			html, err := CreateDiagram(td.code)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if strings.TrimSpace(html) == "" {
				t.Fatal("expected non-empty SVG output")
			}

			for _, fragment := range td.expectedFragments {
				if !strings.Contains(html, fragment) {
					t.Fatalf("expected SVG to contain fragment %q", fragment)
				}
			}
		})
	}
}

func TestDatabaseCoordinatesAreNotAppliedTwice(t *testing.T) {
	code := `
db:Database
@db(x:100,y:50,w:200,h:120)
`

	html, err := CreateDiagram(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(html, `transform="translate(100.000000,50.000000)"`) {
		t.Fatalf("expected outer render-tree translation for database node")
	}

	if !strings.Contains(html, `transform="translate(0, 0)"`) {
		t.Fatalf("expected database component template to render in local coordinates")
	}
}
