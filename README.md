<h1 align="center">Content Maestro</h1>

<div align="center">

![License](https://img.shields.io/github/license/think-root/content-maestro?style=flat-square&color=blue)
[![Go Report Card](https://goreportcard.com/badge/github.com/think-root/content-maestro?style=flat-square)](https://goreportcard.com/report/github.com/think-root/content-maestro)
[![Go Version](https://img.shields.io/github/go-mod/go-version/think-root/content-maestro?style=flat-square&color=blue)](https://github.com/think-root/content-maestro)
[![Deploy Status](https://img.shields.io/github/actions/workflow/status/think-root/content-maestro/deploy.yml?branch=main&label=Deploy&style=flat-square)](https://github.com/think-root/content-maestro/actions/workflows/deploy.yml)
[![Version](https://img.shields.io/github/v/release/think-root/content-maestro?style=flat-square&color=blue)](https://github.com/think-root/content-maestro/releases)
[![Changelog](https://img.shields.io/badge/changelog-view-blue?style=flat-square)](CHANGELOG.md)

<img src="baner.png" alt="baner">

</div>

## Description

Helper app for [content-alchemist](https://github.com/think-root/content-alchemist) that manages content like a skilled maestro. Essentially, it makes scheduled requests to various integrations (such as [telegram-connector](https://github.com/think-root/telegram-connector) or [x-connector](https://github.com/think-root/x-connector)) using a convenient [config](internal/api/apis-config.yml). It also prepares posts for publication by generating images with information about the repository using [socialify](https://github.com/wei/socialify) and makes scheduled requests to the API method of [content-alchemist](https://github.com/think-root/content-alchemist), which [automatically generates](https://github.com/think-root/content-alchemist?tab=readme-ov-file#apiauto-generate) new posts.

### SQLite Migration (v3.0.0+)

Starting from version **3.0.0**, the application migrated from PostgreSQL to **SQLite** for simpler deployment and reduced infrastructure requirements. The migration is handled automatically by [`internal/store/migration.go`](internal/store/migration.go) — on startup, the app detects if a PostgreSQL instance is available and seamlessly migrates all data (cron settings, history, collect settings, prompt configuration) to the local SQLite database. Once migration completes, a flag is set to prevent re-migration on subsequent restarts. No manual intervention required — just deploy and the app takes care of the rest.

### Technology Stack

- Go 1.24
- SQLite 3.24

## Install

### Requirements

- [Go 1.24](https://go.dev/doc/install) or later

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

| Variable                  | Required                     | Description |
| ------------------------- | ---------------------------- | ----------- |
| API_TOKEN                 | Yes                          | Bearer token checked by the API middleware; requests return an error without it. |
| API_PORT                  | No (default: 8080)           | Port for the API server. |
| SQLITE_DB_PATH            | No (default: ./data/content-maestro.db) | Path to SQLite database file. |
| CONTENT_ALCHEMIST_URL     | Yes                          | Base URL for content-alchemist endpoints used by collectors and message jobs. |
| CONTENT_ALCHEMIST_BEARER  | Yes                          | Bearer token for content-alchemist requests. |
| CONTENT_ALCHEMIST_TIMEOUT | No (default: 300s collect / 30s repo) | Timeout in seconds for content-alchemist calls. |
| TWITTER_URL               | Yes (enabled by default)     | URL of the Twitter/X connector server. |
| TWITTER_API_KEY           | Yes (enabled by default)     | API key header for the Twitter/X connector. |
| TELEGRAM_SERVER_URL       | Yes (enabled by default)     | URL of the Telegram connector server. |
| TELEGRAM_SERVER_TOKEN     | Yes (enabled by default)     | API key header for the Telegram connector. |
| BLUESKY_URL               | Yes (enabled by default)     | URL of the Bluesky connector server. |
| BLUESKY_SERVER_KEY        | Yes (enabled by default)     | API key header for the Bluesky connector. |
| WAPP_SERVER_URL           | Only if enabling WhatsApp    | URL of the WhatsApp connector server (disabled by default in `apis-config.yml`). |
| WAPP_TOKEN                | Only if enabling WhatsApp    | API key for the WhatsApp connector. |
| WAPP_JID                  | Only if enabling WhatsApp    | Target WhatsApp chat/channel JID for `/wapp/send-message`. |
| PUBLIC_URL                | Yes (for Threads)            | Base URL (e.g., https://yourdomain.com) for serving images to external APIs. |

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

## External APIs Integration

Content Maestro integrates with various external platforms (Twitter/X, Telegram, Bluesky, WhatsApp). API configurations are now managed through the REST API endpoints, stored in the SQLite database.

### Managing API Configurations

API configurations can be managed through the following endpoints:

- `GET /api/api-configs` - Retrieve all API configurations
- `GET /api/api-configs/{name}` - Retrieve a specific API configuration
- `POST /api/api-configs` - Create a new API configuration
- `PUT /api/api-configs/{name}` - Update an existing API configuration
- `DELETE /api/api-configs/{name}` - Delete an API configuration

All requests require the `Authorization: Bearer <API_TOKEN>` header.

### Configuration Fields

Each API configuration contains the following fields:

- `name`: Unique identifier for the API configuration
- `url`: The endpoint URL with environment variable support (uses `{env.VAR}` syntax)
- `method`: HTTP method for the request (GET, POST, PUT, DELETE, PATCH)
- `auth_type`: Authentication type ("bearer", "api_key", or empty)
- `token_env_var`: Environment variable name containing the auth token
- `token_header`: Header name for API key (if auth_type is "api_key")
- `content_type`: Request content type ("json" or "multipart")
- `timeout`: Request timeout in seconds (must be > 0)
- `success_code`: Expected HTTP success response code (100-599)
- `enabled`: Boolean flag to enable/disable the API
- `response_type`: Expected response format
- `text_language`: Optional language code for text content (e.g., "en", "uk")
- `socialify_image`: Boolean flag to enable/disable socialify image generation for this API
- `default_json_body`: JSON string of key/value pairs always added to requests (supports `{env.VAR}` interpolation)

### Migration from YAML

On first startup (v3.5.0+), the application automatically migrates existing API configurations from `apis-config.yml` to the SQLite database. After migration, the YAML file is no longer used for runtime configuration but can be kept as a backup. All subsequent configuration changes should be made through the REST API endpoints.

## Application API

Content Maestro exposes its own REST API for managing jobs, schedules, and settings. For detailed documentation of the application's API endpoints, authentication, and usage examples, see [API Documentation](api_docs.md).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
