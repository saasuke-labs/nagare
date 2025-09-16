import { makeScene2D, Rect, Line, Txt, Circle } from "@motion-canvas/2d"
import {
  createRef,
  all,
  waitFor,
  tween,
  easeInOutCubic,
} from "@motion-canvas/core"
import { Browser, Server, VM } from "../../components"

const BROWSER_LEFT = -400
const BROWSER_WIDTH = 350
const BROWSER_RIGHT = BROWSER_LEFT + BROWSER_WIDTH

const VM_LEFT = 350
const VM_WIDTH = 450
const VM_RIGHT = VM_LEFT + VM_WIDTH

const NGINX_LEFT = VM_LEFT + 100
const NGINX_WIDTH = 120
const NGINX_RIGHT = NGINX_LEFT + NGINX_WIDTH

const APP_LEFT = NGINX_RIGHT + 100
const APP_WIDTH = 120
const APP_RIGHT = APP_LEFT + APP_WIDTH

const connect = (from: Rect, to: Rect, ref: Line) => {
  const fromPos = from.position()
  const toPos = to.position()

  const fromSize = from.size()
  const toSize = to.size()

  const startX = fromPos.x + fromSize.x / 2
  const startY = fromPos.y + fromSize.y / 2

  const endX = toPos.x - toSize.x / 2
  const endY = toPos.y + toSize.y / 2

  ref.points([
    [startX, startY],
    [endX, endY],
  ])
}

export default makeScene2D(function* (view) {
  // References
  const browser = createRef<Browser>()
  const vm = createRef<VM>()
  const nginx = createRef<Server>()
  const appServer = createRef<Server>()

  // Request flow elements
  const requestArrow = createRef<Line>()
  const nginxArrow = createRef<Line>()
  const responseArrow = createRef<Line>()
  const dataPacket = createRef<Circle>()
  const responseText = createRef<Txt>()

  // Layout
  view.add(
    <>
      <Browser
        ref={browser}
        x={BROWSER_LEFT}
        y={0}
        width={BROWSER_WIDTH}
        height={250}
        title="Client"
        url="http://localhost/status"
        titleBarColor="#2d71eeff"
      />

      <VM
        ref={vm}
        x={VM_LEFT}
        y={0}
        width={VM_WIDTH}
        height={300}
        vmName="VPS"
        status="stopped"
        cpuUsage={0.2}
        memoryUsage={0.4}
      />

      {/* Request flow arrows */}
      <Line
        ref={requestArrow}
        lineWidth={3}
        stroke="#3b82f6"
        endArrow
        opacity={0}
        lineDash={[8, 4]}
      />

      <Line
        ref={nginxArrow}
        lineWidth={2}
        stroke="#10b981"
        points={[
          [NGINX_RIGHT, 0], // NGINX position
          [APP_LEFT, -60], // App server position
        ]}
        endArrow
        opacity={0}
        lineDash={[6, 3]}
      />

      <Line
        ref={responseArrow}
        lineWidth={3}
        stroke="#f59e0b"
        points={[
          [NGINX_LEFT, 0], // VM edge (response path)
          [BROWSER_RIGHT, 0], // Browser edge (response path)
        ]}
        endArrow
        opacity={0}
        lineDash={[8, 4]}
      />

      {/* Data packet */}
      <Circle ref={dataPacket} size={12} fill="#3b82f6" opacity={0} />

      {/* Response text */}
      <Rect
        ref={responseText}
        fill="#1f2937"
        stroke="#10b981"
        lineWidth={2}
        radius={6}
        padding={12}
        x={BROWSER_LEFT}
        y={0}
        opacity={0}
      >
        <Txt
          text='{"status": "ok"}'
          fill="#10b981"
          fontSize={16}
          fontWeight={600}
          fontFamily="monospace"
        />
      </Rect>
    </>
  )

  connect(browser(), vm(), requestArrow())

  // Add components to VM
  vm().addComponent(
    <Server
      ref={nginx}
      label="NGINX"
      serverType="nginx"
      status="active"
      width={NGINX_WIDTH}
      height={50}
      x={NGINX_LEFT}
    />,
    -100,
    -60
  )

  vm().addComponent(
    <Server
      ref={appServer}
      label="APP"
      serverType="app"
      status="active"
      width={APP_WIDTH}
      height={50}
      fill="#3b82f6"
      x={APP_LEFT}
    />,
    100,
    -60
  )

  // === Animation Sequence ===

  // 1. Initial setup - show browser and VM
  yield* all(browser().open(0.5), vm().bootUp(0.8))

  yield* waitFor(0.5)

  // 2. Start request flow
  yield* browser().url("http://localhost/api/status", 0.3)

  // 3. Show request arrow and animate data packet
  yield* requestArrow().opacity(1, 0.3)

  dataPacket().position([-225, 0])
  yield* dataPacket().opacity(1, 0.2)

  // 4. Animate packet from browser to VM
  yield* dataPacket().position([125, 0], 0.8)

  // 5. Packet enters VM, highlight NGINX
  yield* nginx().shadowBlur(15, 0.3)
  yield* nginx().shadowColor("#22c55e", 0.3)

  // 6. Show NGINX to app arrow
  yield* nginxArrow().opacity(1, 0.3)

  // 7. Move packet from NGINX to app server
  dataPacket().position([250, -60])
  yield* dataPacket().position([450, -60], 0.5)

  // 8. App server processes request
  yield* appServer().shadowBlur(15, 0.3)
  yield* appServer().shadowColor("#3b82f6", 0.3)

  yield* waitFor(0.3)

  // 9. Response comes back - change packet color to indicate response
  yield* dataPacket().fill("#f59e0b", 0.2)

  // 10. Move response back to NGINX
  yield* dataPacket().position([250, -60], 0.5)

  // 11. Clear app server highlight
  yield* appServer().shadowBlur(0, 0.3)

  // 12. NGINX forwards response back
  dataPacket().position([125, 20]) // Move to response path
  yield* responseArrow().opacity(1, 0.3)

  // 13. Response travels back to browser
  yield* dataPacket().position([-225, 20], 0.8)

  // 14. Clear NGINX highlight
  yield* nginx().shadowBlur(0, 0.3)

  // 15. Show response in browser
  yield* all(
    dataPacket().opacity(0, 0.2),
    responseText().opacity(1, 0.5),
    browser().url("http://localhost/api/status ✓", 0.3)
  )

  yield* waitFor(1)

  // 16. Hide response and reset for loop
  yield* all(
    responseText().opacity(0, 0.3),
    requestArrow().opacity(0, 0.3),
    nginxArrow().opacity(0, 0.3),
    responseArrow().opacity(0, 0.3)
  )

  yield* waitFor(0.5)
})
