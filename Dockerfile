# Stage 1: Build (builder)
FROM golang:1.24 AS builder
WORKDIR /app
COPY src/* ./
RUN go mod download
RUN go build

# Stage 2: Runtime
FROM debian:bookworm-slim AS workspace
RUN apt update && apt install ca-certificates -y
WORKDIR /app
COPY --from=builder /app/squidarr-proxy /app/squidarr-proxy
RUN mkdir -p /data/squidarr
ENTRYPOINT ["./squidarr-proxy"]
