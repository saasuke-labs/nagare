package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"log"
	"math"
	"net/http"
	"strings"

	"nagare-go/anim"
	"nagare-go/mermaid"
	"nagare-go/model"

	"github.com/fogleman/gg"
)

// --------- Scene JSON (payload) ---------

//type KfN = anim.KfN

// type Node struct {
// 	ID    string  `json:"id"`
// 	Label string  `json:"label,omitempty"`
// 	X     float64 `json:"x"`
// 	Y     float64 `json:"y"`
// 	W     float64 `json:"w"`
// 	H     float64 `json:"h"`

// 	XTrack  []KfN `json:"xTrack,omitempty"`
// 	YTrack  []KfN `json:"yTrack,omitempty"`
// 	Opacity []KfN `json:"opacity,omitempty"`
// }

// type Edge struct {
// 	ID     string `json:"id"`
// 	From   string `json:"from"`
// 	To     string `json:"to"`
// 	Label  string `json:"label,omitempty"`
// 	FlowOn []KfN  `json:"flowOn,omitempty"`
// }

// type Scene struct {
// 	Width       int     `json:"width,omitempty"`  // default 960
// 	Height      int     `json:"height,omitempty"` // default 540
// 	FPS         float64 `json:"fps,omitempty"`    // default 15
// 	DurationSec float64 `json:"durationSec"`      // required

// 	Nodes []Node `json:"nodes"`
// 	Edges []Edge `json:"edges"`

// 	Bg string `json:"bg,omitempty"` // e.g. "#ffffff"
// 	Fg string `json:"fg,omitempty"` // e.g. "#222222"
// }

// --------- small color helpers ---------

func parseHexRGB(s string) (r, g, b float64, ok bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, 0, 0, false
	}
	if s[0] == '#' {
		s = s[1:]
	}
	switch len(s) {
	case 3: // #rgb
		r8, _ := hex.DecodeString(strings.Repeat(string(s[0]), 2))
		g8, _ := hex.DecodeString(strings.Repeat(string(s[1]), 2))
		b8, _ := hex.DecodeString(strings.Repeat(string(s[2]), 2))
		return float64(r8[0]) / 255, float64(g8[0]) / 255, float64(b8[0]) / 255, true
	case 6: // #rrggbb
		bs, err := hex.DecodeString(s)
		if err != nil || len(bs) != 3 {
			return 0, 0, 0, false
		}
		return float64(bs[0]) / 255, float64(bs[1]) / 255, float64(bs[2]) / 255, true
	default:
		return 0, 0, 0, false
	}
}
func setHexRGBA(dc *gg.Context, hex string, a float64) {
	if r, g, b, ok := parseHexRGB(hex); ok {
		dc.SetRGBA(r, g, b, a)
		return
	}
	dc.SetRGBA(0, 0, 0, a) // fallback
}
func colorFromHex(hexStr string) color.Color {
	if r, g, b, ok := parseHexRGB(hexStr); ok {
		return color.RGBA{uint8(r*255 + 0.5), uint8(g*255 + 0.5), uint8(b*255 + 0.5), 255}
	}
	return color.White
}

// --------- HTTP server ---------

func main() {
	http.HandleFunc("/render/gif", handleGIF)

	http.HandleFunc("/render/mermaid/gif", handleMermaidGIF)
	log.Println("listening on :8080  (POST /render/gif)")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleMermaidGIF(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request Received")
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		fmt.Println("bad json: " + err.Error())

		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Data Code: ", data.Code)

	g, err := mermaid.ParseFlowchart(data.Code)

	if err != nil {
		fmt.Println("bad mermaid: " + err.Error())
		http.Error(w, "bad mermaid: "+err.Error(), http.StatusBadRequest)
		return
	}

	dot := g.ToDOT()
	L, err := mermaid.LayoutWithDotPlain(dot)
	if err != nil {
		fmt.Println("bad mermaid2: " + err.Error())
		http.Error(w, "bad mermaid2: "+err.Error(), http.StatusBadRequest)
		return
	}

	scene := mermaid.BuildScene(g, L, 15, 3.0, 960, 540)

	buf, err := toGIF(scene)

	if err != nil {
		fmt.Println("bad json: " + err.Error())
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("content-type", "image/gif")
	w.WriteHeader(200)
	_, _ = w.Write(buf.Bytes())

}

func toGIF(scene model.Scene) (bytes.Buffer, error) {

	fmt.Println("Here Start")

	if scene.FPS <= 0 {
		scene.FPS = 15
	}
	if scene.Width <= 0 {
		scene.Width = 960
	}
	if scene.Height <= 0 {
		scene.Height = 540
	}
	if scene.Bg == "" {
		scene.Bg = "#ffffff"
	}
	if scene.Fg == "" {
		scene.Fg = "#222222"
	}
	if scene.DurationSec <= 0 {
		return bytes.Buffer{}, fmt.Errorf("scene.DurationSec <= 0")
	}

	total := int(math.Round(scene.FPS * scene.DurationSec))
	var frames []*image.Paletted
	var delays []int

	for i := 0; i < total; i++ {
		t := float64(i) / scene.FPS
		dc := gg.NewContext(scene.Width, scene.Height)

		// background
		bg := colorFromHex(scene.Bg)
		dc.SetColor(bg)
		dc.Clear()

		fmt.Println("Here Mid")

		// edges first
		for _, e := range scene.Edges {
			from := nodeByID(scene.Nodes, e.From)
			to := nodeByID(scene.Nodes, e.To)
			if from == nil || to == nil {
				continue
			}

			fx := from.X + at(from.XTrack, t) + from.W
			fy := from.Y + at(from.YTrack, t) + from.H/2
			tx := to.X + at(to.XTrack, t)
			ty := to.Y + at(to.YTrack, t) + to.H/2
			mx := (fx + tx) / 2

			setHexRGBA(dc, scene.Fg, 1)
			dc.SetLineWidth(2.5)
			dc.NewSubPath()
			dc.MoveTo(fx, fy)
			dc.LineTo(mx, fy)
			dc.LineTo(mx, ty)
			dc.LineTo(tx, ty)

			on := anim.AtNumber(e.FlowOn, t)
			if on >= 0.5 {
				dc.SetDash(14, 10)
			} else {
				dc.SetDash()
			}
			dc.Stroke()
			dc.SetDash()
		}

		// nodes
		for _, n := range scene.Nodes {
			alpha := clamp01(anim.AtNumber(n.Opacity, t))
			x := n.X + at(n.XTrack, t)
			y := n.Y + at(n.YTrack, t)

			dc.Push()
			// fill
			dc.SetRGBA(0.95, 0.97, 1.0, alpha)
			dc.DrawRoundedRectangle(x, y, n.W, n.H, 14)
			dc.FillPreserve()

			// stroke
			setHexRGBA(dc, scene.Fg, alpha)
			dc.SetLineWidth(2)
			dc.Stroke()
			dc.Pop()

			// label
			if n.Label != "" {
				setHexRGBA(dc, scene.Fg, alpha)
				dc.DrawStringAnchored(n.Label, x+n.W/2, y+n.H/2, 0.5, 0.5)
			}
		}

		// quantize quickly (for best quality, build a palette once from all frames)
		img := dc.Image()
		b := img.Bounds()
		pal := image.NewPaletted(b, palette.WebSafe)
		draw.FloydSteinberg.Draw(pal, b, img, image.Point{})
		frames = append(frames, pal)
		delays = append(delays, int(100.0/scene.FPS))
	}

	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, &gif.GIF{Image: frames, Delay: delays, LoopCount: 0}); err != nil {
		return bytes.Buffer{}, err
	}
	return buf, nil
}

func handleGIF(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request Received")
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var scene model.Scene
	if err := json.NewDecoder(r.Body).Decode(&scene); err != nil {
		fmt.Println("bad json: " + err.Error())

		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}

	buf, err := toGIF(scene)

	if err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Returing GIF")
	w.Header().Set("content-type", "image/gif")
	w.WriteHeader(200)
	_, _ = w.Write(buf.Bytes())
}

func nodeByID(nodes []model.SceneNode, id string) *model.SceneNode {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i]
		}
	}
	return nil
}

func at(track []model.KfN, t float64) float64 { return anim.AtNumber(track, t) }
func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
