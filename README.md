<h1 align="center">Content Maestro</h1>

<div align="center">

![License](https://img.shields.io/github/license/think-root/content-maestro?style=flat-square&color=blue)
[![Go Report Card](https://goreportcard.com/badge/github.com/think-root/content-maestro?style=flat-square)](https://goreportcard.com/report/github.com/think-root/content-maestro)
[![Go Version](https://img.shields.io/github/go-mod/go-version/think-root/content-maestro?style=flat-square&color=blue)](https://github.com/think-root/content-maestro)
[![Deploy Status](https://img.shields.io/github/actions/workflow/status/think-root/content-maestro/deploy.yml?branch=main&label=Deploy&style=flat-square)](https://github.com/think-root/content-maestro/actions/workflows/deploy.yml)
[![Version](https://img.shields.io/github/v/release/think-root/content-maestro?style=flat-square&color=blue)](https://github.com/think-root/content-maestro/releases)
[![Changelog](https://img.shields.io/badge/changelog-view-blue?style=flat-square)](CHANGELOG.md)

<!-- ![Coverage](https://img.shields.io/badge/Coverage-28%25-red.svg) -->

<img src="baner.png" alt="baner">

</div>

## Description

Helper app for [content-alchemist](https://github.com/think-root/content-alchemist) that manages content like a skilled maestro. Essentially, it makes scheduled requests to various integrations (such as [telegram-connector](https://github.com/think-root/telegram-connector) or [x-connector](https://github.com/think-root/x-connector)) using a convenient [config](internal/api/apis-config.yml). It also prepares posts for publication by generating images with information about the repository using [socialify](https://github.com/wei/socialify) and makes scheduled requests to the API method of [content-alchemist](https://github.com/think-root/content-alchemist), which [automatically generates](https://github.com/think-root/content-alchemist?tab=readme-ov-file#apiauto-generate) new posts.

### Technology Stack

- Go 1.24
- PostgreSQL

## Install

### Requirements

- [Go 1.24](https://go.dev/doc/install) or later
- [PostgreSQL 17](https://www.postgresql.org/download/)

### Clone repo

```shell
git clone https://github.com/think-root/content-maestro.git
cd content-maestro
```

### Install dependencies

```bash
go mod download
```

### Config

Create a **.env** file in the app root directory:

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
| API_TOKEN                | Authentication token for the API server                                                                   |
| API_PORT                 | Port for the API server (default: 8080)                                                                   |
| POSTGRES_HOST            | PostgreSQL database host (default: localhost)                                                             |
| POSTGRES_PORT            | PostgreSQL database port (default: 5432)                                                                  |
| POSTGRES_USER            | PostgreSQL database username                                                                               |
| POSTGRES_PASSWORD        | PostgreSQL database password                                                                               |
| POSTGRES_DB              | PostgreSQL database name                                                                                   |

> [!WARNING]
> WhatsApp integration is unofficial and may risk account suspension

### Run the app

```bash
go run ./cmd/main.go
```

The API will be accessible at `http://localhost:8080` (or the port specified in `API_PORT`).

### Building (optional)

```bash
go build -o content-maestro ./cmd/main.go
./content-maestro
```

## APIs integration config

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

## API Documentation

For detailed API documentation, including endpoints, authentication, and usage examples, see [API Documentation](api_docs.md).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
