# Build
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG APP_VERSION=dev
RUN go build -ldflags="-X 'content-maestro/config.APP_VERSION=${APP_VERSION}'" -o content-maestro ./cmd/main.go

# Runtime
FROM alpine:3.16
WORKDIR /app
COPY --from=builder /app/content-maestro .
COPY .env /app/.env
COPY assets/ /app/assets/
CMD ["./content-maestro"]