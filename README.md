<h1 align="center">Content Maestro</h1>

<div align="center">

![License](https://img.shields.io/github/license/think-root/content-maestro?style=flat-square)
[![Go Version](https://img.shields.io/github/go-mod/go-version/think-root/content-maestro)](https://github.com/think-root/content-maestro)
[![Version](https://img.shields.io/github/v/release/think-root/content-maestro)](https://github.com/think-root/content-maestro/releases)
[![Changelog](https://img.shields.io/badge/changelog-view-blue)](CHANGELOG.md)
[![Deploy Status](https://github.com/think-root/content-maestro/workflows/Deploy%20content-maestro/badge.svg)](https://github.com/think-root/content-maestro/actions/workflows/deploy.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/think-root/content-maestro)](https://goreportcard.com/report/github.com/think-root/content-maestro)
![Coverage](https://img.shields.io/badge/Coverage-49%25-red.svg)

<img src="baner.png" alt="baner">

</div>

## Description

Helper app for [content-alchemist](https://github.com/think-root/content-alchemist) that manages content like a skilled maestro. Essentially, it makes scheduled requests to various integrations (such as [telegram-connector](https://github.com/think-root/telegram-connector) or [x-connector](https://github.com/think-root/x-connector)) using a convenient [config](internal/api/apis-config.yml). It also prepares posts for publication by generating images with information about the repository using [socialify](https://github.com/wei/socialify) and makes scheduled requests to the API method of [content-alchemist](https://github.com/think-root/content-alchemist), which [automatically generates](https://github.com/think-root/content-alchemist?tab=readme-ov-file#apiauto-generate) new posts.

### Technology Stack

- Go 1.24
- Docker & Docker Compose

## How to run

### Requirements

- [docker](https://docs.docker.com/engine/install/) or/and [docker-compose](https://docs.docker.com/compose/install/)

### Clone repo

```shell
git clone https://github.com/think-root/content-maestro.git
```

### Config

create a **.env** file in the app root directory

| Variable                 | Description                                                                                               |
| ------------------------ | --------------------------------------------------------------------------------------------------------- |
| CONTENT_ALCHEMIST_BEARER | The token you created when deploying [content-alchemist](https://github.com/think-root/content-alchemist) |
| CONTENT_ALCHEMIST_URL    | The URL of the content-alchemist server, e.g., http://localhost:8080                                      |
| TWITTER_API_KEY          | Your API key for integration with [Twitter](https://github.com/think-root/x-connector)                    |
| TWITTER_URL              | The server URL, e.g., http://localhost:8080                                                               |
| WAPP_TOKEN               | Your API key for integration with [WhatsApp](https://github.com/think-root/whatsapp-connector)            |
| WAPP_JID                 | WhatsApp Channel ID                                                                                       |
| WAPP_SERVER_URL          | The URL of the WhatsApp integration server, e.g., http://localhost:8080                                   |
| TELEGRAM_SERVER_URL      | The URL of the Telegram integration server, e.g., http://localhost:8080                                   |
| TELEGRAM_SERVER_TOKEN    | Your API key for integration with [Telegram](https://github.com/think-root/telegram-connector)            |

⚠️ Warning: WhatsApp integration is unofficial and may risk account suspension


## Apis config

The [apis-config.yml](internal/api/apis-config.yml) file contains configuration settings for various messaging APIs used by the content-maestro service.

### Structure

Each API configuration contains the following fields:

- `url`: The endpoint URL with environment variable support
- `method`: HTTP method for the request
- `auth_type`: Authentication type ("bearer" or "api_key")
- `token_env_var`: Environment variable name containing the auth token
- `token_header`: Header name for API key (if auth_type is "api_key")
- `content_type`: Request content type ("json" or "multipart")
- `timeout`: Request timeout in seconds
- `success_code`: Expected HTTP success response code
- `enabled`: Boolean flag to enable/disable the API
- `response_type`: Expected response format

### Supported APIs

Currently configured APIs:

**WhatsApp**

- Uses bearer token authentication
- JSON content type
- Currently disabled by default

**Twitter**

- Uses API key authentication via X-API-Key header
- Multipart content type
- Enabled by default

**Telegram**

- Uses API key authentication via X-API-Key header
- Multipart content type
- Enabled by default

### Environment Variables Required

- `WAPP_SERVER_URL`
- `WAPP_TOKEN`
- `TWITTER_URL`
- `TWITTER_API_KEY`
- `TELEGRAM_SERVER_URL`
- `TELEGRAM_SERVER_TOKEN`

### Deploy

```bash
docker compose up -d
```

## Contribution

### Development Setup

1. Install Go 1.24 or later
2. Clone the repository
3. Install dependencies: `go mod download`

### Running Locally

1. Set up your .env file
2. Run the app:
  ```bash
  go run ./cmd/main.go
  ```

### Building

```bash
go build -o content-maestro ./cmd/main.go
```

### Testing

```bash
go test -v -cover ./...
```

### Pull Requests

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests to ensure everything works
5. Commit your changes (`git commit -m 'Add some amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Guidelines

- Follow Go coding standards and conventions
- Include tests for new features
- Update documentation as needed
- Keep commits atomic and well-described
- Reference issues in commit messages and PRs
