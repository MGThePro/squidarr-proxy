FROM golang:1.24

WORKDIR /app

COPY src/* ./
RUN go mod download

RUN go build


RUN mkdir /data
RUN mkdir /data/squidarr

ENTRYPOINT ["./squidarr-proxy"]
