FROM golang:1.25.2-bookworm as builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl ca-certificates git build-essential pkg-config && \
    rm -rf /var/lib/apt/lists/*

RUN apt-get update && apt-get install -y nocache git ca-certificates && update-ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/gateway cmd/gatewaysvc/main.go


FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && \
    rm -rf /var/lib/apt/lists/*

RUN useradd -ms /bin/bash --uid 1000 gateway
WORKDIR /app
COPY --from=builder /app/bin /app
COPY --from=builder /app/static ./static
EXPOSE 8080
EXPOSE 50051
USER gateway

CMD ["./gateway"]