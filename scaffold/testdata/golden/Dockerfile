# ── build stage ───────────────────────────────────────────────────────────────
FROM golang:1.25.1-alpine AS builder

WORKDIR /src

# Cache module downloads separately from source compilation.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w \
      -X github.com/nojyerac/go-lib/version.semVer=$(cat VERSION 2>/dev/null || echo 0.0.0) \
      -X github.com/nojyerac/go-lib/version.gitSHA=$(git rev-list -1 HEAD 2>/dev/null || echo unknown)" \
    -o /out/example \
    ./cmd/example

# ── runtime stage ─────────────────────────────────────────────────────────────
FROM scratch

COPY --from=builder /out/example /example

# Default: no-TLS, port 8080.  Override at runtime via env vars.
ENV EXAMPLE_NO_TLS=true \
    EXAMPLE_PORT=8080 \
    EXAMPLE_LOG_LEVEL=info \
    EXAMPLE_EXPORTER_TYPE=noop \
    EXAMPLE_HEALTHCHECK_CHECK_INTERVAL=30s

EXPOSE 8080

ENTRYPOINT ["/example"]
