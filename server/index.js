const express = require("express")
const { parseMermaid } = require("./mermaid.js")
const path = require("path")
const { render } = require("./render.js")
const fs = require("fs").promises
const app = express()

const PORT = process.env.PORT || 3000

// Middleware to parse JSON bodies
app.use(express.json())

// Endpoint to parse Mermaid diagrams
app.post("/parse-mermaid", async (req, res) => {
  console.log("Body:", req.body)
  const { code } = req.body
  if (!code) {
    return res.status(400).json({ error: "No Mermaid code provided" })
  }

  try {
    const result = await parseMermaid(code)

    await fs.writeFile(
      path.join(__dirname, "motion", "data.json"),
      JSON.stringify(result, null, 2)
    )

    await render(
      path.join(__dirname, "motion", "project.ts"),
      path.join(__dirname, "animation.mp4")
    )
    console.log("Parsed result:", result)
    res.json(result)
  } catch (error) {
    console.error("Error parsing Mermaid code:", error)
    res.status(500).json({ error: "Failed to parse Mermaid code" })
  }
})

app.listen(PORT, () => {
  console.log(`Server is running on http://localhost:${PORT}`)
})
