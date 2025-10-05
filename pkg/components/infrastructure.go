package components

import (
	"fmt"

	"github.com/saasuke-labs/nagare/pkg/props"
)

// DatabaseProps defines configurable values for a database component.
type DatabaseProps struct {
	Title           string `prop:"title"`
	Engine          string `prop:"engine"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	AccentColor     string `prop:"accent"`
}

func (d *DatabaseProps) Parse(input string) error {
	return props.ParseProps(input, d)
}

func DefaultDatabaseProps() DatabaseProps {
	return DatabaseProps{
		Title:           "Database",
		Engine:          "PostgreSQL",
		BackgroundColor: "#0f766e",
		ForegroundColor: "#ecfdf5",
		AccentColor:     "#14b8a6",
	}
}

type Database struct {
	Shape
	Text  string
	Props DatabaseProps
	State string
}

func NewDatabase(id string) *Database {
	return &Database{
		Text:  id,
		Props: DefaultDatabaseProps(),
	}
}

type DatabaseTemplateData struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
	Props  DatabaseProps
	Text   string
}

func (d *Database) templateData() DatabaseTemplateData {
	return DatabaseTemplateData{
		X:      d.X,
		Y:      d.Y,
		Width:  d.Width,
		Height: d.Height,
		Props:  d.Props,
		Text:   d.Text,
	}
}

func (d *Database) Draw() string {
	result, err := RenderTemplate("database", d.templateData())
	if err != nil {
		return fmt.Sprintf("<!-- Error rendering database template: %v -->", err)
	}
	return result
}

// MessageQueueProps defines configurable values for a message queue component.
type MessageQueueProps struct {
	Title           string `prop:"title"`
	Kind            string `prop:"kind"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	AccentColor     string `prop:"accent"`
}

func (m *MessageQueueProps) Parse(input string) error {
	return props.ParseProps(input, m)
}

func DefaultMessageQueueProps() MessageQueueProps {
	return MessageQueueProps{
		Title:           "Queue",
		Kind:            "RabbitMQ",
		BackgroundColor: "#4c1d95",
		ForegroundColor: "#ede9fe",
		AccentColor:     "#a855f7",
	}
}

type MessageQueue struct {
	Shape
	Text  string
	Props MessageQueueProps
	State string
}

func NewMessageQueue(id string) *MessageQueue {
	return &MessageQueue{
		Text:  id,
		Props: DefaultMessageQueueProps(),
	}
}

type MessageQueueTemplateData struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
	Props  MessageQueueProps
	Text   string
}

func (m *MessageQueue) templateData() MessageQueueTemplateData {
	return MessageQueueTemplateData{
		X:      m.X,
		Y:      m.Y,
		Width:  m.Width,
		Height: m.Height,
		Props:  m.Props,
		Text:   m.Text,
	}
}

func (m *MessageQueue) Draw() string {
	result, err := RenderTemplate("message-queue", m.templateData())
	if err != nil {
		return fmt.Sprintf("<!-- Error rendering message queue template: %v -->", err)
	}
	return result
}

// CDNProps defines configurable values for an edge or CDN component.
type CDNProps struct {
	Title           string `prop:"title"`
	Provider        string `prop:"provider"`
	Region          string `prop:"region"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	AccentColor     string `prop:"accent"`
}

func (c *CDNProps) Parse(input string) error {
	return props.ParseProps(input, c)
}

func DefaultCDNProps() CDNProps {
	return CDNProps{
		Title:           "Edge",
		Provider:        "Cloudflare",
		Region:          "Global",
		BackgroundColor: "#1d4ed8",
		ForegroundColor: "#eff6ff",
		AccentColor:     "#60a5fa",
	}
}

type CDN struct {
	Shape
	Text  string
	Props CDNProps
	State string
}

func NewCDN(id string) *CDN {
	return &CDN{
		Text:  id,
		Props: DefaultCDNProps(),
	}
}

type CDNTemplateData struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
	Props  CDNProps
	Text   string
}

func (c *CDN) templateData() CDNTemplateData {
	return CDNTemplateData{
		X:      c.X,
		Y:      c.Y,
		Width:  c.Width,
		Height: c.Height,
		Props:  c.Props,
		Text:   c.Text,
	}
}

func (c *CDN) Draw() string {
	result, err := RenderTemplate("cdn", c.templateData())
	if err != nil {
		return fmt.Sprintf("<!-- Error rendering CDN template: %v -->", err)
	}
	return result
}

// APIGatewayProps defines configurable values for an API Gateway component.
type APIGatewayProps struct {
	Title           string `prop:"title"`
	Route           string `prop:"route"`
	Method          string `prop:"method"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	AccentColor     string `prop:"accent"`
}

func (a *APIGatewayProps) Parse(input string) error {
	return props.ParseProps(input, a)
}

func DefaultAPIGatewayProps() APIGatewayProps {
	return APIGatewayProps{
		Title:           "API Gateway",
		Route:           "/api",
		Method:          "ANY",
		BackgroundColor: "#312e81",
		ForegroundColor: "#eef2ff",
		AccentColor:     "#6366f1",
	}
}

type APIGateway struct {
	Shape
	Text  string
	Props APIGatewayProps
	State string
}

func NewAPIGateway(id string) *APIGateway {
	return &APIGateway{
		Text:  id,
		Props: DefaultAPIGatewayProps(),
	}
}

type APIGatewayTemplateData struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
	Props  APIGatewayProps
	Text   string
}

func (a *APIGateway) templateData() APIGatewayTemplateData {
	return APIGatewayTemplateData{
		X:      a.X,
		Y:      a.Y,
		Width:  a.Width,
		Height: a.Height,
		Props:  a.Props,
		Text:   a.Text,
	}
}

func (a *APIGateway) Draw() string {
	result, err := RenderTemplate("api-gateway", a.templateData())
	if err != nil {
		return fmt.Sprintf("<!-- Error rendering API gateway template: %v -->", err)
	}
	return result
}

// BackgroundWorkerProps defines configurable values for a worker component.
type BackgroundWorkerProps struct {
	Title           string `prop:"title"`
	Job             string `prop:"job"`
	Schedule        string `prop:"schedule"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	AccentColor     string `prop:"accent"`
}

func (b *BackgroundWorkerProps) Parse(input string) error {
	return props.ParseProps(input, b)
}

func DefaultBackgroundWorkerProps() BackgroundWorkerProps {
	return BackgroundWorkerProps{
		Title:           "Worker",
		Job:             "process-emails",
		Schedule:        "@every 1m",
		BackgroundColor: "#166534",
		ForegroundColor: "#dcfce7",
		AccentColor:     "#22c55e",
	}
}

type BackgroundWorker struct {
	Shape
	Text  string
	Props BackgroundWorkerProps
	State string
}

func NewBackgroundWorker(id string) *BackgroundWorker {
	return &BackgroundWorker{
		Text:  id,
		Props: DefaultBackgroundWorkerProps(),
	}
}

type BackgroundWorkerTemplateData struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
	Props  BackgroundWorkerProps
	Text   string
}

func (b *BackgroundWorker) templateData() BackgroundWorkerTemplateData {
	return BackgroundWorkerTemplateData{
		X:      b.X,
		Y:      b.Y,
		Width:  b.Width,
		Height: b.Height,
		Props:  b.Props,
		Text:   b.Text,
	}
}

func (b *BackgroundWorker) Draw() string {
	result, err := RenderTemplate("background-worker", b.templateData())
	if err != nil {
		return fmt.Sprintf("<!-- Error rendering background worker template: %v -->", err)
	}
	return result
}

// PackageProps defines configurable values for a package artifact component.
type PackageProps struct {
	Title           string `prop:"title"`
	Version         string `prop:"version"`
	Language        string `prop:"lang"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	AccentColor     string `prop:"accent"`
}

func (p *PackageProps) Parse(input string) error {
	return props.ParseProps(input, p)
}

func DefaultPackageProps() PackageProps {
	return PackageProps{
		Title:           "Package",
		Version:         "1.0.0",
		Language:        "Go",
		BackgroundColor: "#92400e",
		ForegroundColor: "#fef3c7",
		AccentColor:     "#f97316",
	}
}

type Package struct {
	Shape
	Text  string
	Props PackageProps
	State string
}

func NewPackage(id string) *Package {
	return &Package{
		Text:  id,
		Props: DefaultPackageProps(),
	}
}

type PackageTemplateData struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
	Props  PackageProps
	Text   string
}

func (p *Package) templateData() PackageTemplateData {
	return PackageTemplateData{
		X:      p.X,
		Y:      p.Y,
		Width:  p.Width,
		Height: p.Height,
		Props:  p.Props,
		Text:   p.Text,
	}
}

func (p *Package) Draw() string {
	result, err := RenderTemplate("package", p.templateData())
	if err != nil {
		return fmt.Sprintf("<!-- Error rendering package template: %v -->", err)
	}
	return result
}

// ArtifactProps defines configurable values for a file or artifact component.
type ArtifactProps struct {
	Title           string `prop:"title"`
	Filename        string `prop:"filename"`
	Size            string `prop:"size"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	AccentColor     string `prop:"accent"`
}

func (a *ArtifactProps) Parse(input string) error {
	return props.ParseProps(input, a)
}

func DefaultArtifactProps() ArtifactProps {
	return ArtifactProps{
		Title:           "Artifact",
		Filename:        "build.tar.gz",
		Size:            "24 MB",
		BackgroundColor: "#1f2937",
		ForegroundColor: "#f9fafb",
		AccentColor:     "#9ca3af",
	}
}

type Artifact struct {
	Shape
	Text  string
	Props ArtifactProps
	State string
}

func NewArtifact(id string) *Artifact {
	return &Artifact{
		Text:  id,
		Props: DefaultArtifactProps(),
	}
}

type ArtifactTemplateData struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
	Props  ArtifactProps
	Text   string
}

func (a *Artifact) templateData() ArtifactTemplateData {
	return ArtifactTemplateData{
		X:      a.X,
		Y:      a.Y,
		Width:  a.Width,
		Height: a.Height,
		Props:  a.Props,
		Text:   a.Text,
	}
}

func (a *Artifact) Draw() string {
	result, err := RenderTemplate("artifact", a.templateData())
	if err != nil {
		return fmt.Sprintf("<!-- Error rendering artifact template: %v -->", err)
	}
	return result
}
