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

func TestDatabaseActionsMapToComposedLEDs(t *testing.T) {
	code := `
	db:Database
	@db(x:100,y:50,w:200,h:120)
	@db.read(begin:"0s",dur:"1s")
	@db.write(begin:"1s",dur:"1.5s")
	`

	html, err := CreateDiagram(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(html, `values="darkgreen;#4ade80;#22c55e;#86efac;#16a34a;#4ade80;darkgreen"`) {
		t.Fatalf("expected read action to map to green led blink animation")
	}

	if !strings.Contains(html, `values="darkred;#f87171;#ef4444;#dc2626;#b91c1c;#f87171;darkred"`) {
		t.Fatalf("expected write action to map to red led blink animation")
	}
}

func TestPrimitiveCylinderAndLedRender(t *testing.T) {
	code := `
	c:Cylinder
	l:Led
	@c(x:10,y:20,w:100,h:80,title:"Body",subtitle:"Storage")
	@l(x:140,y:30,mode:"red")
	@l.blink(begin:"0s",dur:"2s")
	`

	html, err := CreateDiagram(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(html, `Storage`) {
		t.Fatalf("expected cylinder subtitle text")
	}
	if !strings.Contains(html, `darkred;#f87171;#ef4444;#dc2626;#b91c1c;#f87171;darkred`) {
		t.Fatalf("expected red led blink palette in SVG")
	}
}

func TestDatabaseLedLayoutIsInsideAndScalesWithDatabase(t *testing.T) {
	small := `
	db:Database
	@db(x:100,y:50,w:100,h:100)
	@db.read(begin:"0s",dur:"1s")
	@db.write(begin:"0s",dur:"1s")
	`

	large := `
	db:Database
	@db(x:100,y:50,w:300,h:300)
	@db.read(begin:"0s",dur:"1s")
	@db.write(begin:"0s",dur:"1s")
	`

	smallSVG, err := CreateDiagram(small)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	largeSVG, err := CreateDiagram(large)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(smallSVG, `transform="translate(66.5, 10)"`) {
		t.Fatalf("expected small database leds to stay inside right edge")
	}
	if !strings.Contains(largeSVG, `transform="translate(199.5, 30)"`) {
		t.Fatalf("expected large database leds to move proportionally with larger size")
	}
	if !strings.Contains(smallSVG, `rx="5" ry="5"`) {
		t.Fatalf("expected smaller led radius for smaller database")
	}
	if !strings.Contains(largeSVG, `rx="15" ry="15"`) {
		t.Fatalf("expected larger led radius for larger database")
	}
}
