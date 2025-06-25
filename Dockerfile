# Stage 1: Build (builder)
FROM golang:1.24 AS builder
WORKDIR /app
COPY src/* ./
RUN go build -o squidarr-proxy

# Stage 2: Runtime
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/squidarr-proxy /app/
RUN mkdir -p /data/squidarr
ENTRYPOINT ["./squidarr-proxy"]
