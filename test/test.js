const test = async (code) => {
  const res = await fetch("http://localhost:3000/parse-mermaid", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      code,
    }),
  })

  if (!res.ok) {
    console.error("Error:", await res.text())
    process.exit(1)
  }

  const data = await res.json()
  console.log("Response:", JSON.stringify(data, null, 2))
}

// test(`flowchart LR
//       B[Browser https://example.com]:::Browser --> S[Server api.example.com]:::Server
//       S --> D[(DB orders)]:::DB
//       B -."url@1.5:https://app.example.com/dashboard".-> B
//       D -."highlight@2.2".-> D

//     classDef Browser fill:#fff,stroke:#222;
//     classDef Server fill:#fff,stroke:#222;
//     classDef DB fill:#fff,stroke:#222;
//   `).then(() => process.exit(0))

// test(`flowchart LR
//       B[Browser https://example.com]:::Browser

//     classDef Browser fill:#fff,stroke:#222;
//   `).then(() => process.exit(0))

test(`flowchart LR
      B[Browser https://example.com]:::Browser 
      VM:::VM 

      B -- "request" --> VM
    
    classDef Browser fill:#fff,stroke:#222;
    classDef VM fill:#333,stroke:#ccc;
  `).then(() => process.exit(0))
