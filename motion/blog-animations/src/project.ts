import { makeProject } from "@motion-canvas/core"

import example from "./scenes/example?scene"
import bluegreen from "./scenes/bluegreen/bluegreen?scene"
import bluegreenWithComponents from "./scenes/bluegreen-with-components?scene"
import browserExample from "./scenes/browser-example?scene"
import fromjson from "./scenes/fromjson?scene"

export default makeProject({
  scenes: [
    //bluegreen,
    fromjson,
    //bluegreenWithComponents,
    //browserExample,
  ],
})
