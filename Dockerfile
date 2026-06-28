# Multi-stage build: compile the Go binary, then run it in a tiny image.

# --- Stage 1: build ---
FROM golang:1.25-alpine AS build
WORKDIR /src
# Cache dependencies first (only re-runs when go.mod/go.sum change).
COPY go.mod go.sum* ./
RUN go mod download
# Copy the rest and build a static binary.
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api

# --- Stage 2: run ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/api /app/api
COPY migrations /app/migrations
EXPOSE 8080
ENTRYPOINT ["/app/api"]
