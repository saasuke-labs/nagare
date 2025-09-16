const puppeteer = require("puppeteer")

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms))

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
      const label = node.querySelector("text")?.textContent
      const rect = node.querySelector("rect, ellipse")
      let bbox = null
      if (rect && typeof rect.getBBox === "function") {
        const b = rect.getBBox()
        bbox = { x: b.x, y: b.y, width: b.width, height: b.height }
      }
      return { id, label, bbox }
    })
  )

  // Extract edges
  const edges = await page.$$eval("g.edge", (edges) =>
    edges.map((edge) => {
      const id = edge.id
      const label = edge.querySelector("text")?.textContent
      // Optionally, extract source/target from path or marker
      return { id, label }
    })
  )

  await browser.close()
  return { nodes, edges }
}

module.exports = { parseMermaid }
