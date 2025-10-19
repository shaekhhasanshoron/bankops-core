FROM golang:1.25.2-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl ca-certificates git build-essential pkg-config && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /app/bin/auth cmd/txsvc/main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && \
    rm -rf /var/lib/apt/lists/*

RUN useradd -ms /bin/bash --uid 1000 transaction
WORKDIR /app
RUN mkdir -p /data && chown -R transaction:transaction /data /app
COPY --from=builder /app/bin/auth /app/transaction

EXPOSE 8080
EXPOSE 50051

USER transaction

CMD ["/app/transaction"]