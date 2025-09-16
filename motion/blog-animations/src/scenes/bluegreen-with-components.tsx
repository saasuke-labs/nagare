import { makeScene2D, Line, Txt, Circle } from "@motion-canvas/2d"
import {
  createRef,
  all,
  waitFor,
  tween,
  easeInOutCubic,
} from "@motion-canvas/core"
import { Server } from "../components"

export default makeScene2D(function* (view) {
  // References
  const client = createRef<Server>()
  const nginx = createRef<Server>()
  const blue = createRef<Server>()
  const green = createRef<Server>()
  const arrowBlue = createRef<Line>()
  const arrowGreen = createRef<Line>()
  const pulse = createRef<Circle>()
  const caption = createRef<Txt>()

  // Layout using reusable Server components
  view.add(
    <>
      <Server
        ref={client}
        x={-500}
        y={0}
        label="Client"
        serverType="client"
        status="active"
      />

      <Server
        ref={nginx}
        x={-200}
        y={0}
        width={140}
        label="NGINX"
        serverType="nginx"
        status="active"
      />

      <Server
        ref={blue}
        x={200}
        y={-80}
        label="Blue"
        serverType="app"
        status="active"
        fill="#3b82f6"
      />

      <Server
        ref={green}
        x={200}
        y={80}
        label="Green"
        serverType="app"
        status="inactive"
        fill="#22c55e"
        opacity={0.2}
      />

      <Line
        ref={arrowBlue}
        lineWidth={4}
        stroke="#3b82f6"
        points={[
          [-130, 0],
          [140, -80],
        ]}
        endArrow
      />
      <Line
        ref={arrowGreen}
        lineWidth={4}
        stroke="#22c55e"
        points={[
          [-130, 0],
          [140, 80],
        ]}
        endArrow
        opacity={0}
      />

      <Circle ref={pulse} size={16} fill="white" x={-500} y={0} opacity={0} />

      <Txt
        ref={caption}
        y={180}
        x={-250}
        text="current: blue"
        fill="white"
        fontSize={24}
      />
    </>
  )

  // === Timeline ===

  // Fade in
  yield* all(
    client().opacity(1, 0.3),
    nginx().opacity(1, 0.3),
    blue().opacity(1, 0.3)
  )

  // Idle to blue (0.0–0.3)
  pulse().opacity(1)
  yield* tween(0.3, (v) => {
    pulse().position.x(-200 + v * 340) // client(-500) → nginx(-200) → blue(200)
    pulse().position.y(-v * 80) // curve to blue
  })

  // Spin up green (0.3–0.8)
  yield* all(
    green().activate(0.5),
    green().deploy(0.5),
    caption().text("deploy green...", 0.5)
  )

  // Switch to green (0.8–1.2)
  yield* all(
    arrowBlue().opacity(0, 0.4),
    arrowGreen().opacity(1, 0.4),
    caption().text("switch traffic", 0.4)
  )

  // Serve green (1.2–1.8)
  yield* tween(0.6, (v) => {
    pulse().position.x(-200 + v * 340)
    pulse().position.y(v * 80)
  })
  yield* caption().text("current: green", 0.2)

  // Prepare blue (1.8–2.3)
  yield* all(caption().text("next deploy: blue...", 0.5), blue().deploy(0.5))

  // Switch back (2.3–2.7)
  yield* all(
    arrowGreen().opacity(0, 0.4),
    arrowBlue().opacity(1, 0.4),
    caption().text("switch traffic", 0.4)
  )

  // Serve blue (2.7–3.0)
  yield* tween(0.3, (v) => {
    pulse().position.x(-200 + v * 340)
    pulse().position.y(-v * 80)
  })
  yield* caption().text("current: blue", 0.2)

  yield* waitFor(0.3) // Hold for loop
})
