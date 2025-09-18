const { spawn } = require("child_process")

async function render(projectPath, outPath) {
  await new Promise((resolve, reject) => {
    const p = spawn(
      "npx",
      [
        "@motion-canvas/cli",
        "render",
        "--project",
        projectPath,
        "--output",
        outPath,
        "--fps",
        "30",
        "--width",
        "1280",
        "--height",
        "720",
      ],
      { stdio: "inherit" }
    )
    p.on("exit", (code) =>
      code === 0 ? resolve() : reject(new Error(`MC render failed: ${code}`))
    )
  })
}

module.exports = { render }
