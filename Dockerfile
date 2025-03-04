# Build
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build content-maestro ./cmd/main.go

# Runtime
FROM alpine:3.16
WORKDIR /app
ARG APP_VERSION
ENV APP_VERSION=${APP_VERSION}
COPY --from=builder /app/content-maestro .
COPY .env /app/.env
COPY assets/ /app/assets/
CMD ["./content-maestro"]