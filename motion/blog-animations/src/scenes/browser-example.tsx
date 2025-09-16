import { makeScene2D, Txt } from "@motion-canvas/2d"
import { createRef, all, waitFor } from "@motion-canvas/core"
import { Browser } from "../components"

export default makeScene2D(function* (view) {
  // Create references to browser instances
  const browser1 = createRef<Browser>()
  const browser2 = createRef<Browser>()
  const browser3 = createRef<Browser>()

  // Add browsers to the scene
  view.add(
    <>
      <Browser
        ref={browser1}
        x={-400}
        y={0}
        width={350}
        height={250}
        title="Development"
        url="http://localhost:3000"
        titleBarColor="#dbeafe"
      />

      <Browser
        ref={browser2}
        x={0}
        y={0}
        width={350}
        height={250}
        title="Staging"
        url="https://staging.example.com"
        titleBarColor="#fef3c7"
      />

      <Browser
        ref={browser3}
        x={400}
        y={0}
        width={350}
        height={250}
        title="Production"
        url="https://example.com"
        titleBarColor="#dcfce7"
      />
    </>
  )

  // Add some content to the browsers
  browser1().addContent(
    <Txt text="Dev Server" fill="#3b82f6" fontSize={24} fontWeight={600} />
  )

  browser2().addContent(
    <Txt text="Staging" fill="#f59e0b" fontSize={24} fontWeight={600} />
  )

  browser3().addContent(
    <Txt text="Live Site" fill="#10b981" fontSize={24} fontWeight={600} />
  )

  // Animation sequence
  yield* browser1().open(0.3)
  yield* waitFor(0.2)
  yield* browser2().open(0.3)
  yield* waitFor(0.2)
  yield* browser3().open(0.3)

  yield* waitFor(2)

  // Demonstrate updating properties
  yield* all(
    browser1().title("Updated Dev", 0.5),
    browser2().url("https://staging-v2.example.com", 0.5),
    browser3().titleBarColor("#fecaca", 0.5)
  )

  yield* waitFor(1)

  // Close all browsers
  yield* all(
    browser1().close(0.3),
    browser2().close(0.3),
    browser3().close(0.3)
  )
})
