import {
  Rect,
  RectProps,
  Txt,
  Circle,
  Line,
  Node,
  Layout,
  initial,
  signal,
} from "@motion-canvas/2d"
import { SignalValue, SimpleSignal, easeInOutCubic } from "@motion-canvas/core"
import { Server } from "./Server"

export interface VMProps extends RectProps {
  vmName?: SignalValue<string>
  status?: SignalValue<"running" | "stopped" | "starting" | "deploying">
  cpuUsage?: SignalValue<number>
  memoryUsage?: SignalValue<number>
  showMetrics?: SignalValue<boolean>
}

export class VM extends Rect {
  @initial("VM-001")
  @signal()
  public declare readonly vmName: SimpleSignal<string, this>

  @initial("running")
  @signal()
  public declare readonly status: SimpleSignal<
    "running" | "stopped" | "starting" | "deploying",
    this
  >

  @initial(0.3)
  @signal()
  public declare readonly cpuUsage: SimpleSignal<number, this>

  @initial(0.6)
  @signal()
  public declare readonly memoryUsage: SimpleSignal<number, this>

  @initial(true)
  @signal()
  public declare readonly showMetrics: SimpleSignal<boolean, this>

  private header: Rect
  private contentArea: Layout
  private statusIndicator: Circle
  private vmNameText: Txt
  private statusText: Txt
  private cpuBar: Rect
  private memoryBar: Rect
  private cpuFill: Rect
  private memoryFill: Rect
  private cpuLabel: Txt
  private memoryLabel: Txt
  private borderLines: Line[]

  public constructor(props?: VMProps) {
    super({
      fill: "#1f2937",
      stroke: "#374151",
      lineWidth: 2,
      radius: 12,
      width: 500,
      height: 350,
      shadowColor: "#00000040",
      shadowBlur: 15,
      shadowOffsetY: 5,
      ...props,
    })

    // Header section
    this.header = new Rect({
      fill: "#111827",
      width: () => this.width() - 4,
      height: 50,
      radius: [10, 10, 0, 0],
      y: () => -this.height() / 2 + 25,
      stroke: "#374151",
      lineWidth: 1,
    })

    // Status indicator
    this.statusIndicator = new Circle({
      size: 12,
      fill: () => this.getStatusColor(),
      x: () => -this.width() / 2 + 25,
      y: () => -this.height() / 2 + 25,
    })

    // VM name
    this.vmNameText = new Txt({
      text: this.vmName,
      fill: "#f9fafb",
      fontSize: 16,
      fontWeight: 600,
      x: () => -this.width() / 2 + 50,
      y: () => -this.height() / 2 + 25,
    })

    // Status text
    this.statusText = new Txt({
      text: () => this.status().toUpperCase(),
      fill: () => this.getStatusColor(),
      fontSize: 12,
      fontWeight: 500,
      x: () => this.width() / 2 - 60,
      y: () => -this.height() / 2 + 25,
    })

    // Content area for components
    this.contentArea = new Layout({
      y: () => this.height() / 2 - (this.height() - 100) / 2,
      width: () => this.width() - 40,
      height: () => this.height() - 100,
    })

    // Metrics section
    const metricsY = () => this.height() / 2 - 40

    // CPU usage bar
    this.cpuBar = new Rect({
      fill: "#374151",
      width: 120,
      height: 8,
      radius: 4,
      x: () => -this.width() / 2 + 80,
      y: metricsY,
    })

    this.cpuFill = new Rect({
      fill: "#3b82f6",
      width: () => 120 * this.cpuUsage(),
      height: 8,
      radius: 4,
      x: () => -this.width() / 2 + 80 - (120 - 120 * this.cpuUsage()) / 2,
      y: metricsY,
    })

    this.cpuLabel = new Txt({
      text: () => `CPU: ${Math.round(this.cpuUsage() * 100)}%`,
      fill: "#9ca3af",
      fontSize: 10,
      x: () => -this.width() / 2 + 80,
      y: () => metricsY() - 15,
    })

    // Memory usage bar
    this.memoryBar = new Rect({
      fill: "#374151",
      width: 120,
      height: 8,
      radius: 4,
      x: () => this.width() / 2 - 80,
      y: metricsY,
    })

    this.memoryFill = new Rect({
      fill: "#10b981",
      width: () => 120 * this.memoryUsage(),
      height: 8,
      radius: 4,
      x: () => this.width() / 2 - 80 - (120 - 120 * this.memoryUsage()) / 2,
      y: metricsY,
    })

    this.memoryLabel = new Txt({
      text: () => `RAM: ${Math.round(this.memoryUsage() * 100)}%`,
      fill: "#9ca3af",
      fontSize: 10,
      x: () => this.width() / 2 - 80,
      y: () => metricsY() - 15,
    })

    // Decorative corner accents (simpler approach)
    this.borderLines = [
      // Top left
      new Line({
        lineWidth: 3,
        stroke: "#3b82f6",
        points: [
          [-this.width() / 2 + 15, -this.height() / 2],
          [-this.width() / 2, -this.height() / 2],
          [-this.width() / 2, -this.height() / 2 + 15],
        ],
      }),
      // Top right
      new Line({
        lineWidth: 3,
        stroke: "#3b82f6",
        points: [
          [this.width() / 2 - 15, -this.height() / 2],
          [this.width() / 2, -this.height() / 2],
          [this.width() / 2, -this.height() / 2 + 15],
        ],
      }),
      // Bottom left
      new Line({
        lineWidth: 3,
        stroke: "#3b82f6",
        points: [
          [-this.width() / 2 + 15, this.height() / 2],
          [-this.width() / 2, this.height() / 2],
          [-this.width() / 2, this.height() / 2 - 15],
        ],
      }),
      // Bottom right
      new Line({
        lineWidth: 3,
        stroke: "#3b82f6",
        points: [
          [this.width() / 2 - 15, this.height() / 2],
          [this.width() / 2, this.height() / 2],
          [this.width() / 2, this.height() / 2 - 15],
        ],
      }),
    ]

    // Add all elements
    this.add([
      this.header,
      this.statusIndicator,
      this.vmNameText,
      this.statusText,
      this.contentArea,
      ...this.borderLines,
    ])

    // Add metrics if enabled
    this.add([
      this.cpuBar,
      this.cpuFill,
      this.cpuLabel,
      this.memoryBar,
      this.memoryFill,
      this.memoryLabel,
    ])

    // Hide metrics initially if showMetrics is false
    if (!this.showMetrics()) {
      this.cpuBar.opacity(0)
      this.cpuFill.opacity(0)
      this.cpuLabel.opacity(0)
      this.memoryBar.opacity(0)
      this.memoryFill.opacity(0)
      this.memoryLabel.opacity(0)
    }
  }

  private getStatusColor(): string {
    switch (this.status()) {
      case "running":
        return "#10b981"
      case "stopped":
        return "#ef4444"
      case "starting":
        return "#f59e0b"
      case "deploying":
        return "#8b5cf6"
      default:
        return "#6b7280"
    }
  }

  // Method to add components to the VM
  public addComponent(component: Node, x: number = 0, y: number = 0) {
    component.position([x, y])
    this.contentArea.add(component)
    return this
  }

  // Animation methods
  public *bootUp(duration: number = 1.0) {
    this.status("starting")
    this.scale(0.8)
    this.opacity(0)

    yield* this.opacity(1, duration * 0.3)
    yield* this.scale(1, duration * 0.7, easeInOutCubic)
    yield* this.status("running", 0)
  }

  public *shutdown(duration: number = 0.8) {
    yield* this.status("stopped", 0)
    yield* this.scale(0.8, duration * 0.6, easeInOutCubic)
    yield* this.opacity(0, duration * 0.4)
  }

  public *showMetricsAnimation(duration: number = 0.5) {
    yield* this.showMetrics(true, 0)
    yield* this.cpuBar.opacity(1, duration)
    yield* this.cpuFill.opacity(1, duration)
    yield* this.cpuLabel.opacity(1, duration)
    yield* this.memoryBar.opacity(1, duration)
    yield* this.memoryFill.opacity(1, duration)
    yield* this.memoryLabel.opacity(1, duration)
  }

  public *updateCpuUsage(newUsage: number, duration: number = 0.3) {
    yield* this.cpuUsage(newUsage, duration)
  }

  public *updateMemoryUsage(newUsage: number, duration: number = 0.3) {
    yield* this.memoryUsage(newUsage, duration)
  }

  public *deployMode(duration: number = 0.5) {
    yield* this.status("deploying", 0)
    // Pulsing effect during deployment
    yield* this.borderLines[0].stroke("#8b5cf6", duration)
    yield* this.borderLines[1].stroke("#8b5cf6", duration)
    yield* this.borderLines[2].stroke("#8b5cf6", duration)
    yield* this.borderLines[3].stroke("#8b5cf6", duration)
  }

  public *normalMode(duration: number = 0.3) {
    yield* this.status("running", 0)
    yield* this.borderLines[0].stroke("#3b82f6", duration)
    yield* this.borderLines[1].stroke("#3b82f6", duration)
    yield* this.borderLines[2].stroke("#3b82f6", duration)
    yield* this.borderLines[3].stroke("#3b82f6", duration)
  }
}
