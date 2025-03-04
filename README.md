<h1 align="center">Content Maestro</h1>

[![Go Version](https://img.shields.io/github/go-mod/go-version/think-root/content-maestro)](https://github.com/think-root/content-maestro)
[![License](https://img.shields.io/github/license/think-root/content-maestro)](LICENSE)
[![Version](https://img.shields.io/github/v/release/think-root/content-maestro)](https://github.com/think-root/content-maestro/releases)
[![Changelog](https://img.shields.io/badge/changelog-view-blue)](CHANGELOG.md)
[![Deploy Status](https://github.com/think-root/content-maestro/workflows/Deploy%20content-maestro/badge.svg)](https://github.com/think-root/content-maestro/actions/workflows/deploy.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/think-root/content-maestro)](https://goreportcard.com/report/github.com/think-root/content-maestro)

<div align="center">
    <img src="baner.png" alt="baner">
</div>

## Description

This app is a part of [content-alchemist](https://github.com/think-root/content-alchemist), essentially an app that schedules requests to various integrations, which then publish content.

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
