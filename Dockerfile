# Build stage
FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o server ./cmd/server


# Runtime stage
FROM alpine:3.22

RUN addgroup -S app && adduser -S -G app app

WORKDIR /

COPY --from=builder /app/server /server
COPY --from=builder /app/migrations /migrations

USER app:app

EXPOSE 8080

CMD ["./server"]
