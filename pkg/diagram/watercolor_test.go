package diagram

import (
	"strings"
	"testing"

	"github.com/saasuke-labs/nagare/pkg/components"
)

const watercolorBody = `
shop:Browser(url: "https://shop.example.com", text: "Checkout", x:30,y:120,w:230,h:170)
gw:APIGateway(title: "gateway", method: "POST", route: "/orders", x:310,y:&shop.c,w:180,h:60)
vm:VM(title: "prod@ubuntu", x:540,y:40,w:430,h:200) {
    api:Server(title: "orders-api", icon: "golang", port: 8080, x:30,y:60,w:170,h:50)
    mailer:BackgroundWorker(title: "mailer", job: "send-receipt", x:240,y:60,w:160,h:50)
}
db:Database(title: "orders", engine: "postgres", x:570,y:300,w:120,h:120)
events:MessageQueue(title: "order.created", kind: "nats", x:770,y:325,w:200,h:70)
ci:Terminal(title: "ci", cwd: "~/site", command: "go test", x:20,y:330,w:240,h:110)

@shop.request(target: @gw)
@gw.request(target: @api)
@api.request(target: @db)
@mailer.response(target: @events)
@db.write(begin:"0.5s",dur:"1.5s")
`

func renderWatercolor(t *testing.T, layout string) string {
	t.Helper()
	components.ResetArrowMarkerCounter()
	svg, err := CreateDiagram(layout + "\n" + watercolorBody)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return svg
}

func TestWatercolorIsDeterministic(t *testing.T) {
	a := renderWatercolor(t, "@layout(w:1000,h:450,style:watercolor)")
	b := renderWatercolor(t, "@layout(w:1000,h:450,style:watercolor)")
	if a != b {
		t.Fatal("expected the same diagram to paint identically every time")
	}
}

func TestWatercolorPaintsEveryComponent(t *testing.T) {
	svg := renderWatercolor(t, "@layout(w:1000,h:450,style:watercolor)")

	for _, want := range []string{
		`class="nagare-watercolor"`,
		`id="nagare-wc-paper"`,
		`id="nagare-wc-wash0"`,
		`class="nagare-paper"`,
		`mix-blend-mode:multiply`,
		`class="nagare-ink-arrow"`,
		`class="nagare-led"`,
		`<animate attributeName="fill"`, // the database light still blinks
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("expected painted SVG to contain %q", want)
		}
	}
	for _, class := range []string{"nagare-browser", "nagare-apigateway", "nagare-vm", "nagare-server", "nagare-backgroundworker", "nagare-cylinder", "nagare-messagequeue", "nagare-terminal"} {
		start := strings.Index(svg, class)
		if start < 0 {
			t.Errorf("missing component %s", class)
			continue
		}
		if !strings.Contains(svg[start:], `class="nagare-wash`) {
			t.Errorf("expected %s to be painted with a wash", class)
		}
	}
	// Flat primitives must not leak into the painted look.
	if strings.Contains(svg, "<polyline") || strings.Contains(svg, `marker-end=`) {
		t.Error("expected ink arrows, found flat polyline arrows")
	}
}

func TestWatercolorPaperCanBeLeftOut(t *testing.T) {
	svg := renderWatercolor(t, "@layout(w:1000,h:450,style:watercolor,paper:none)")
	if strings.Contains(svg, `class="nagare-paper"`) {
		t.Fatal("expected paper:none to keep the canvas transparent")
	}
	if !strings.Contains(svg, `class="nagare-wash`) {
		t.Fatal("expected the components to still be painted")
	}
}

func TestWatercolorDarkSheetLightens(t *testing.T) {
	svg := renderWatercolor(t, "@layout(w:1000,h:450,theme:dark,style:watercolor)")
	if !strings.Contains(svg, `mix-blend-mode:screen`) {
		t.Fatal("expected washes on a dark sheet to lighten it (screen)")
	}
	if strings.Contains(svg, `mix-blend-mode:multiply`) {
		t.Fatal("did not expect multiply washes on a dark sheet")
	}
}

func TestWatercolorDoesNotChangeTheFlatLook(t *testing.T) {
	svg := renderWatercolor(t, "@layout(w:1000,h:450)")
	for _, painted := range []string{"nagare-wc-", "nagare-wash", "nagare-watercolor", "mix-blend-mode"} {
		if strings.Contains(svg, painted) {
			t.Errorf("flat output contains painted marker %q", painted)
		}
	}
}

func TestWatercolorBrushMarksDifferPerComponent(t *testing.T) {
	code := `@layout(w:600,h:200,style:watercolor)
a:Server(title: "same", x:10,y:10,w:200,h:60)
b:Server(title: "same", x:300,y:10,w:200,h:60)
`
	svg, err := CreateDiagram(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cards := strings.Split(svg, `class="nagare-card`)
	if len(cards) != 3 {
		t.Fatalf("expected two cards, got %d", len(cards)-1)
	}
	strip := func(s string) string {
		// Drop the translate so only the painted marks are compared.
		if i := strings.Index(s, ">"); i >= 0 {
			return s[i:]
		}
		return s
	}
	if strip(cards[1]) == strip(cards[2]) {
		t.Fatal("expected two identical components to get different brush marks")
	}
}
