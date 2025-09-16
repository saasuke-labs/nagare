import {
  Rect,
  RectProps,
  Txt,
  Circle,
  initial,
  signal,
} from "@motion-canvas/2d"
import { SignalValue, SimpleSignal } from "@motion-canvas/core"

export interface ServerProps extends RectProps {
  label?: SignalValue<string>
  status?: SignalValue<"active" | "inactive" | "deploying">
  serverType?: SignalValue<"client" | "nginx" | "app">
}

export class Server extends Rect {
  @initial("Server")
  @signal()
  public declare readonly label: SimpleSignal<string, this>

  @initial("active")
  @signal()
  public declare readonly status: SimpleSignal<
    "active" | "inactive" | "deploying",
    this
  >

  @initial("app")
  @signal()
  public declare readonly serverType: SimpleSignal<
    "client" | "nginx" | "app",
    this
  >

  private labelText: Txt
  private statusIndicator: Circle

  public constructor(props?: ServerProps) {
    super({
      width: 120,
      height: 60,
      radius: 8,
      ...props,
    })

    // Set colors based on server type and status
    this.fill(() => this.getServerColor())

    this.labelText = new Txt({
      text: this.label,
      fill: "white",
      fontSize: 28,
      fontWeight: 600,
    })

    this.statusIndicator = new Circle({
      size: 12,
      fill: () => this.getStatusColor(),
      x: () => this.width() / 2 - 15,
      y: () => -this.height() / 2 + 15,
    })

    this.add([this.labelText, this.statusIndicator])
  }

  private getServerColor(): string {
    const type = this.serverType()
    const status = this.status()

    if (status === "inactive") {
      return type === "client"
        ? "#64748b"
        : type === "nginx"
        ? "#334155"
        : "#6b7280"
    }

    switch (type) {
      case "client":
        return "#64748b"
      case "nginx":
        return "#334155"
      case "app":
        return status === "deploying" ? "#f59e0b" : "#3b82f6"
      default:
        return "#64748b"
    }
  }

  private getStatusColor(): string {
    switch (this.status()) {
      case "active":
        return "#10b981"
      case "inactive":
        return "#6b7280"
      case "deploying":
        return "#f59e0b"
      default:
        return "#6b7280"
    }
  }

  public *deploy(duration: number = 0.5) {
    const originalStatus = this.status()
    yield* this.status("deploying", 0)
    yield* this.scale(0.95, duration / 2).to(1, duration / 2)
    yield* this.status("active", 0)
  }

  public *activate(duration: number = 0.3) {
    yield* this.status("active", 0)
    yield* this.opacity(1, duration)
  }

  public *deactivate(duration: number = 0.3) {
    yield* this.status("inactive", 0)
    yield* this.opacity(0.3, duration)
  }
}
