# ── Stage 1: build ────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy module definition first so the layer is cached between source changes.
# There is no go.sum because this module has no external dependencies.
COPY go.mod ./

# Copy the rest of the source and build a static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o nagare ./cmd/main.go

# ── Stage 2: runtime ──────────────────────────────────────────────────────────
# The binary is fully static (CGO_ENABLED=0), so no OS packages are required.
FROM scratch

WORKDIR /app

# Copy the compiled binary
COPY --from=builder /app/nagare .

# Copy static web assets (playground UI + examples)
COPY --from=builder /app/static ./static

# Cloud Run injects PORT; the app defaults to 8080 which matches Cloud Run's default.
EXPOSE 8080

CMD ["./nagare"]
