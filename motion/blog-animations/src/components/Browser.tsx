import {
  Rect,
  RectProps,
  Txt,
  Circle,
  Line,
  Node,
  initial,
  signal,
} from "@motion-canvas/2d"
import { SignalValue, SimpleSignal } from "@motion-canvas/core"

export interface BrowserProps extends RectProps {
  title?: SignalValue<string>
  url?: SignalValue<string>
  titleBarColor?: SignalValue<string>
  contentColor?: SignalValue<string>
}

export class Browser extends Rect {
  @initial("Browser")
  @signal()
  public declare readonly title: SimpleSignal<string, this>

  @initial("https://example.com")
  @signal()
  public declare readonly url: SimpleSignal<string, this>

  @initial("#f3f4f6")
  @signal()
  public declare readonly titleBarColor: SimpleSignal<string, this>

  @initial("#ffffff")
  @signal()
  public declare readonly contentColor: SimpleSignal<string, this>

  private titleBar: Rect
  private contentArea: Rect
  private titleText: Txt
  private urlText: Txt
  private closeButton: Circle
  private minimizeButton: Circle
  private maximizeButton: Circle

  public constructor(props?: BrowserProps) {
    super({
      fill: "#e5e7eb",
      radius: 8,
      width: 400,
      height: 300,
      ...props,
    })

    // Title bar
    this.titleBar = new Rect({
      fill: this.titleBarColor,
      width: () => this.width(),
      height: 40,
      radius: [8, 8, 0, 0],
      y: () => -this.height() / 2 + 20,
    })

    // Window controls (close, minimize, maximize)
    this.closeButton = new Circle({
      size: 12,
      fill: "#ef4444",
      x: () => -this.width() / 2 + 20,
      y: () => -this.height() / 2 + 20,
    })

    this.minimizeButton = new Circle({
      size: 12,
      fill: "#f59e0b",
      x: () => -this.width() / 2 + 40,
      y: () => -this.height() / 2 + 20,
    })

    this.maximizeButton = new Circle({
      size: 12,
      fill: "#10b981",
      x: () => -this.width() / 2 + 60,
      y: () => -this.height() / 2 + 20,
    })

    // Title text
    this.titleText = new Txt({
      text: this.title,
      fill: "#374151",
      fontSize: 14,
      fontWeight: 600,
      y: () => -this.height() / 2 + 20,
    })

    // Content area
    this.contentArea = new Rect({
      fill: this.contentColor,
      width: () => this.width() - 4,
      height: () => this.height() - 44,
      radius: [0, 0, 6, 6],
      y: () => this.height() / 2 - (this.height() - 44) / 2 - 2,
    })

    // URL bar
    const urlBar = new Rect({
      fill: "#f9fafb",
      stroke: "#d1d5db",
      lineWidth: 1,
      width: () => this.width() - 40,
      height: 24,
      radius: 4,
      y: () => -this.height() / 2 + 60,
    })

    this.urlText = new Txt({
      text: this.url,
      fill: "#6b7280",
      fontSize: 12,
      y: () => -this.height() / 2 + 60,
    })

    // Add all elements to the browser
    this.add([
      this.titleBar,
      this.closeButton,
      this.minimizeButton,
      this.maximizeButton,
      this.titleText,
      this.contentArea,
      urlBar,
      this.urlText,
    ])
  }

  // Method to add content to the browser
  public addContent(content: Node) {
    this.contentArea.add(content)
    return this
  }

  // Method to animate the browser opening
  public *open(duration: number = 0.5) {
    this.scale(0)
    yield* this.scale(1, duration)
  }

  // Method to animate the browser closing
  public *close(duration: number = 0.3) {
    yield* this.scale(0, duration)
  }
}
