// scripts/render-seq.ts
// Run with: tsx scripts/render-seq.ts --project ./src/project.ts --out ./frames --fps 15 --w 1280 --h 720

import { mkdir } from "fs/promises"
import path from "path"
import { fileURLToPath, pathToFileURL } from "node:url"

// 👉 Adjust these imports to your MC version:
import { Renderer } from "@motion-canvas/renderer" // renderer core
import { ImageSequenceExporter } from "@motion-canvas/renderer" // PNG exporter
import type { Project } from "@motion-canvas/core"

const args = Object.fromEntries(
  process.argv.slice(2).reduce((acc, a, i, arr) => {
    if (a.startsWith("--")) acc.push([a.slice(2), arr[i + 1]])
    return acc
  }, [] as [string, string][])
)

const projectPath = path.resolve(args.project ?? "./src/project.ts")
const outDir = path.resolve(args.out ?? "./frames")
const fps = Number(args.fps ?? 15)
const width = Number(args.w ?? 1280)
const height = Number(args.h ?? 720)

;(async () => {
  // 1) import your project module (ESM)
  const mod = await import(pathToFileURL(projectPath).href)
  const project: Project = mod.default // makeProject(...) export

  // 2) configure renderer (dimensions / fps)
  const renderer = new Renderer({
    project,
    fps,
    width,
    height,
  })

  // 3) choose an exporter (PNG sequence)
  await mkdir(outDir, { recursive: true })
  const exporter = new ImageSequenceExporter({
    directory: outDir,
    pattern: "frame-%05d.png", // frame-00001.png ...
  })

  // 4) render the whole project timeline (or clamp duration in your project/scene)
  await renderer.render(exporter)

  console.log("✅ Frames written to", outDir)
})()
