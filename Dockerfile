# Stage 1: Build (builder)
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY src/* ./
RUN go mod download
RUN go build

# Stage 2: Runtime
FROM alpine:latest AS workspace
WORKDIR /app
COPY --from=builder /app/squidarr-proxy /app/squidarr-proxy
RUN mkdir -p /data/squidarr
ENTRYPOINT ["./squidarr-proxy"]
