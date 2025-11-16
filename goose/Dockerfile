FROM golang:1.24rc1 as goose

WORKDIR /build

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY ./migrations /migrations

ENTRYPOINT ["/go/bin/goose", "-dir", "/migrations", "postgres"]
CMD ["up"]
