const puppeteer = require("puppeteer")

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms))

function extractClassDefs(code) {
  const classDefs = {}
  const regex = /^\s*classDef\s+(\w+)\s+(.+);?$/gm
  let match
  while ((match = regex.exec(code))) {
    classDefs[match[1]] = match[2].trim()
  }
  return classDefs
}

function extractEdgesFromCode(code) {
  // Matches: A -- "label" --> B, A --> B, etc.
  const edgeRegex =
    /^\s*([A-Za-z0-9_]+)\s*[-.]+(?:\s*"([^"]*)"\s*)?[-.]+>\s*([A-Za-z0-9_]+)/gm
  const edges = []
  let match
  while ((match = edgeRegex.exec(code))) {
    edges.push({
      from: match[1],
      to: match[3],
      label: match[2] || null,
    })
  }
  return edges
}

async function parseMermaid(code) {
  const browser = await puppeteer.launch()
  const page = await browser.newPage()

  // HTML template for Mermaid rendering
  const html = `
    <html>
      <head>
        <script type="module">
          import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@${
            require("mermaid/package.json").version
          }/dist/mermaid.esm.min.mjs';
          window.mermaid = mermaid;
        </script>
      </head>
      <body>
        <div class="mermaid" id="container">${code}</div>
        <script>
          window.addEventListener('DOMContentLoaded', async () => {
            await window.mermaid.run();
          });
        </script>
      </body>
    </html>
  `

  await page.setContent(html, { waitUntil: "networkidle0" })
  await page.waitForSelector("svg")
  await sleep(2000) // Wait for rendering to complete
  // Extract nodes
  const nodes = await page.$$eval('g[id^="flowchart-"], g.node', (nodes) =>
    nodes.map((node) => {
      const id = node.id
      const texts = Array.from(node.querySelectorAll("text"))
        .map((t) => t.textContent)
        .filter(Boolean)
      const label = texts.join(" ").trim() || null
      const classList = node.getAttribute("class")?.split(" ") || []
      const rect = node.querySelector("rect, ellipse")
      let bbox = null
      if (rect && typeof rect.getBBox === "function") {
        const b = rect.getBBox()
        bbox = { x: b.x, y: b.y, width: b.width, height: b.height }
      }
      return { id, label, classList, bbox }
    })
  )

  // Extract edges
  const edges = extractEdgesFromCode(code)

  await browser.close()
  const classDefs = extractClassDefs(code)
  return { nodes, edges, classDefs }
}

module.exports = { parseMermaid }
