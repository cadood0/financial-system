# ---- Stage 1: build ----
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /financial-system .

# ---- Stage 2: run ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /financial-system ./financial-system
COPY db/migrations ./db/migrations

EXPOSE 8080
CMD ["./financial-system"]