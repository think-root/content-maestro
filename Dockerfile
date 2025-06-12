# Build
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o content-maestro ./cmd/main.go

# Runtime
FROM alpine:3.16
WORKDIR /app
ARG APP_VERSION
ENV APP_VERSION=${APP_VERSION}
COPY --from=builder /app/content-maestro .
COPY assets/ /app/assets/
COPY internal/api/apis-config.yml /app/internal/api/apis-config.yml

RUN mkdir -p /app/tmp/gh_project_img && \
  chown -R nobody:nobody /app/tmp

USER nobody

# Default port (can be overridden by API_PORT env variable)
EXPOSE 8080

CMD ["./content-maestro"]
