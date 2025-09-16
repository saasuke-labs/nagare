import { makeScene2D, Line, Txt, Circle } from "@motion-canvas/2d"
import {
  createRef,
  all,
  waitFor,
  tween,
  easeInOutCubic,
  chain,
} from "@motion-canvas/core"
import { Browser, Server, VM } from "../components"

export default makeScene2D(function* (view) {
  // References
  const browser = createRef<Browser>()
  const vm = createRef<VM>()
  const nginx = createRef<Server>()
  const blueServer = createRef<Server>()
  const greenServer = createRef<Server>()
  const connectionLine = createRef<Line>()
  const dataPacket = createRef<Circle>()
  const title = createRef<Txt>()

  view.fill("#0f172a") // Dark background for tech feel

  // Layout
  view.add(
    <>
      {/* Title */}
      <Txt
        ref={title}
        text="Browser → VM Architecture"
        fill="#f1f5f9"
        fontSize={32}
        fontWeight={700}
        y={-320}
        opacity={0}
      />

      {/* Browser on the left */}
      <Browser
        ref={browser}
        x={-400}
        y={0}
        width={300}
        height={200}
        title="User Browser"
        url="https://myapp.com"
        titleBarColor="#1e40af"
        scale={0}
      />

      {/* VM on the right */}
      <VM
        ref={vm}
        x={350}
        y={0}
        width={450}
        height={300}
        vmName="Production-VM"
        status="stopped"
        cpuUsage={0.2}
        memoryUsage={0.4}
        scale={0}
      />

      {/* Connection line */}
      <Line
        ref={connectionLine}
        lineWidth={4}
        stroke="#06b6d4"
        points={[
          [-250, 0],
          [100, 0],
        ]}
        endArrow
        opacity={0}
        lineDash={[10, 5]}
      />

      {/* Data packet animation */}
      <Circle
        ref={dataPacket}
        size={12}
        fill="#06b6d4"
        x={-250}
        y={0}
        opacity={0}
      />
    </>
  )

  // === Timeline ===

  // 1. Show title and browser
  yield* all(title().opacity(1, 0.5), browser().scale(1, 0.8, easeInOutCubic))

  // Add content to browser
  browser().addContent(
    <Txt text="Loading..." fill="#64748b" fontSize={18} y={20} />
  )

  yield* waitFor(0.5)

  // 2. Boot up the VM
  yield* vm().bootUp(1.2)

  // 3. Add components to VM
  vm().addComponent(
    <Server
      ref={nginx}
      label="NGINX"
      serverType="nginx"
      status="active"
      width={120}
      height={40}
      y={-60}
    />,
    0,
    -60
  )

  vm().addComponent(
    <Server
      ref={blueServer}
      label="Blue"
      serverType="app"
      status="active"
      fill="#3b82f6"
      width={100}
      height={35}
    />,
    -80,
    20
  )

  vm().addComponent(
    <Server
      ref={greenServer}
      label="Green"
      serverType="app"
      status="inactive"
      fill="#22c55e"
      width={100}
      height={35}
      opacity={0.3}
    />,
    80,
    20
  )

  yield* waitFor(0.3)

  // 4. Show connection
  yield* connectionLine().opacity(1, 0.6)

  // 5. Animate data flow
  yield* dataPacket().opacity(1, 0.2)

  // Data packet travels from browser to VM
  yield* tween(1.0, (v) => {
    dataPacket().position.x(-250 + v * 350)
    dataPacket().position.y(Math.sin(v * Math.PI) * -20) // Arc motion
  })

  yield* dataPacket().opacity(0, 0.2)

  // 6. Show VM metrics
  yield* vm().showMetricsAnimation(0.8)

  // 7. Simulate load increase
  yield* all(
    vm().updateCpuUsage(0.7, 1.0),
    vm().updateMemoryUsage(0.8, 1.0),
    browser().url("https://myapp.com/dashboard", 0.5)
  )

  // Update browser content
  browser().addContent(
    <Txt
      text="Connected!"
      fill="#10b981"
      fontSize={18}
      fontWeight={600}
      y={20}
    />
  )

  yield* waitFor(0.5)

  // 8. Deploy new version (activate green server)
  yield* vm().deployMode(0.5)

  yield* all(
    greenServer().activate(0.8),
    greenServer().deploy(0.8),
    blueServer().deactivate(0.8)
  )

  yield* vm().normalMode(0.3)

  // 9. Show final state
  yield* all(vm().updateCpuUsage(0.4, 0.8), vm().updateMemoryUsage(0.5, 0.8))

  // 10. Multiple data packets for traffic
  const packets = []
  for (let i = 0; i < 3; i++) {
    const packet = new Circle({
      size: 8,
      fill: "#06b6d4",
      x: -250,
      y: 0,
      opacity: 0,
    })
    view.add(packet)
    packets.push(packet)
  }

  // Animate multiple packets
  yield* all(
    ...packets.map((packet, i) =>
      chain(
        waitFor(i * 0.3),
        packet.opacity(1, 0.1),
        tween(0.8, (v) => {
          packet.position.x(-250 + v * 350)
          packet.position.y(Math.sin(v * Math.PI + i) * -15)
        }),
        packet.opacity(0, 0.1)
      )
    )
  )

  yield* waitFor(1)

  // 11. Fade out
  yield* all(
    title().opacity(0, 0.5),
    browser().scale(0, 0.5),
    vm().shutdown(0.8),
    connectionLine().opacity(0, 0.5)
  )
})
