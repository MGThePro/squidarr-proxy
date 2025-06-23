FROM golang:1.24

WORKDIR /app

COPY src/* ./
RUN go mod download

RUN go build

ENTRYPOINT ["./squidarr-proxy"]
