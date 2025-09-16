import { makeScene2D, Rect, Txt } from "@motion-canvas/2d"
import { createRef, all, waitFor, chain } from "@motion-canvas/core"
import { Browser, Server, VM } from "../components"

export default makeScene2D(function* (view) {
  // References
  const browser = createRef<Browser>()
  const vm = createRef<VM>()
  const title = createRef<Txt>()

  view.fill("#0f172a")

  // Layout
  view.add(
    <>
      <Txt
        ref={title}
        text="VM as a Container"
        fill="#f1f5f9"
        fontSize={36}
        fontWeight={700}
        y={-250}
      />

      <Browser
        ref={browser}
        x={-300}
        y={-50}
        width={250}
        height={150}
        title="Client"
        url="https://app.example.com"
        scale={0.8}
      />

      <VM
        ref={vm}
        x={200}
        y={0}
        width={400}
        height={250}
        vmName="Web-Server-VM"
        status="running"
        showMetrics={false}
      />
    </>
  )

  // Add a simple web page to browser
  browser().addContent(
    <Txt text="My App" fill="#3b82f6" fontSize={16} fontWeight={600} />
  )

  // === VM Components Layout ===

  // NGINX (reverse proxy)
  vm().addComponent(
    <Server
      label="NGINX"
      serverType="nginx"
      status="active"
      width={110}
      height={35}
    />,
    0,
    -60
  )

  // Application servers side by side
  vm().addComponent(
    <Server
      label="App-1"
      serverType="app"
      status="active"
      fill="#3b82f6"
      width={90}
      height={35}
    />,
    -70,
    10
  )

  vm().addComponent(
    <Server
      label="App-2"
      serverType="app"
      status="active"
      fill="#3b82f6"
      width={90}
      height={35}
    />,
    70,
    10
  )

  // Database
  vm().addComponent(
    <Rect fill="#8b5cf6" width={120} height={35} radius={6} y={60}>
      <Txt text="Database" fill="white" fontSize={14} fontWeight={600} />
    </Rect>,
    0,
    60
  )

  // === Animation ===

  yield* waitFor(1)

  // Show the concept: everything runs inside the VM
  yield* vm().scale(1.1, 0.5).to(1, 0.5)

  yield* waitFor(0.5)

  // Highlight the isolation
  yield* all(vm().vmName("Isolated Environment", 0.8), vm().shadowBlur(25, 0.8))

  yield* waitFor(2)
})
