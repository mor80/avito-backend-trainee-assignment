FROM golang:1.24rc1 AS builder
ENV GOTOOLCHAIN=go1.24.0

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o reviewer ./cmd/service-reviewer && \
    chmod +x docker-entrypoint.sh

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /src/reviewer /app/reviewer
COPY --from=builder /go/bin/goose /app/goose
COPY --from=builder /src/docker-entrypoint.sh /app/entrypoint.sh
COPY configs ./configs
COPY migrations ./migrations
RUN chmod +x /app/entrypoint.sh /app/goose /app/reviewer

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["/app/reviewer"]
