# ---------- Build stage ----------
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /exporter ./cmd/exporter

# ---------- Final image ----------
FROM gcr.io/distroless/static

COPY --from=builder /exporter /exporter
USER root

ENTRYPOINT ["/exporter"]