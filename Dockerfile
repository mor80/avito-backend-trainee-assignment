FROM golang:1.24rc1 AS builder
ENV GOTOOLCHAIN=go1.24.0

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o reviewer ./cmd/service-reviewer

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /src/reviewer /app/reviewer
COPY configs ./configs
RUN chmod +x /app/reviewer

EXPOSE 8080

ENTRYPOINT ["/app/reviewer"]
