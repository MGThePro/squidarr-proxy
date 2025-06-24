FROM golang:1.24

WORKDIR /app

COPY src/* ./
RUN go mod download

RUN go build


RUN mkdir /data
RUN mkdir /data/squidarr
RUN mkdir /data/squidarr/complete
RUN mkdir /data/squidarr/incomplete
RUN addgroup --system users
RUN adduser --system abc --ingroup users
RUN chown -R abc:users /data
RUN chmod -R 755 /data
USER abc:users

ENTRYPOINT ["./squidarr-proxy"]
