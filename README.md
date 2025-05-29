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
| API_TOKEN                | Authentication token for the API server                             |
| API_PORT                 | Port for the API server (default: 8080)                                                   |

> [!WARNING]
> WhatsApp integration is unofficial and may risk account suspension

## Apis integration config

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

The application can be deployed in several ways:

#### 1. Direct Docker Deployment with Docker Compose

```bash
docker compose up -d
```

This will:

- Create a Docker volume named `content-maestro-data` that persists the BadgerDB data
- Expose the API on the port specified in `API_PORT` (defaults to 8080)
- Map the container's API port to the same port on the host machine

You can then access the API at `http://localhost:$API_PORT` (or `http://localhost:8080` if `API_PORT` is not set).

#### 2. Direct Docker Deployment without Docker Compose

If you want to run the container directly with Docker:

```bash
# Create a volume for data persistence
docker volume create content-maestro-data

# Build the image
docker build -t content-maestro:latest .

# Run the container
docker run -d \
  --name content-maestro \
  --restart always \
  -p ${API_PORT:-8080}:${API_PORT:-8080} \
  -v content-maestro-data:/app/data/badger \
  --env-file .env \
  -e TZ=Europe/Kiev \
  content-maestro:latest
```

#### 3. Local Development to Docker Migration

If you've been running the application locally and want to move to Docker while preserving your data:

1. Stop the local application
2. Create a Docker volume:

```bash
docker volume create content-maestro-data
```

3. Copy your local BadgerDB data to the Docker volume:

```bash
# Assuming your local data is in ./data/badger
docker run --rm -v $(pwd)/data/badger:/source -v content-maestro-data:/dest alpine cp -r /source/. /dest/
```

4. Deploy using either Docker Compose or direct Docker commands as shown above.

Your existing cron settings will be preserved in the Docker deployment, and the API will be accessible at the configured port.

## API Reference

The application provides a comprehensive REST API for managing cron jobs, repository collection settings, and cron job history. This API enables you to control scheduled tasks, view and update cron schedules, manage job statuses, configure repository collection parameters, and retrieve cron job execution history with advanced pagination and sorting capabilities.

### Cron Job History

#### Get Cron Job History

```http
GET /api/cron-history
```

Retrieve the history of cron job executions with pagination, sorting, and filtering.

**Query parameters:**

- `name` (optional): Filter by cron job name (`collect` or `message`)
- `page` (optional): Page number (default: 1)
- `limit` (optional): Number of records per page (default: 20)
- `sort` (optional): Sort order by execution date (`asc` or `desc`, default: `desc`)
- `success` (optional): Filter by execution status (`true` or `false`)
- `start_date` (optional): Filter records from this date onwards (format: `YYYY-MM-DD` or RFC3339)
- `end_date` (optional): Filter records up to this date (format: `YYYY-MM-DD` or RFC3339)

**Response format:**

The API returns a paginated response with the following structure:

- `data`: Array of cron history records
- `pagination`: Pagination metadata object containing:
  - `total_count`: Total number of records matching the filters
  - `current_page`: Current page number
  - `total_pages`: Total number of pages available
  - `has_next`: Boolean indicating if there's a next page
  - `has_previous`: Boolean indicating if there's a previous page

**Response example:**

```json
{
  "data": [
    {
      "name": "collect",
      "timestamp": "2024-03-15T10:00:00Z",
      "success": true,
      "output": "Successfully collected 5 repositories"
    },
    {
      "name": "message",
      "timestamp": "2024-03-15T10:05:00Z",
      "success": false,
      "output": "Network error"
    }
  ],
  "pagination": {
    "total_count": 25,
    "current_page": 1,
    "total_pages": 2,
    "has_next": true,
    "has_previous": false
  }
}
```

**Example Usage:**

1. Get all history (default: newest first, 20 records per page):

```bash
curl -H "Authorization: Bearer your_api_token" \
  http://localhost:8080/api/cron-history
```

2. Get history for specific job with custom pagination:

```bash
curl -H "Authorization: Bearer your_api_token" \
  "http://localhost:8080/api/cron-history?name=collect&page=1&limit=10"
```

3. Get only failed executions sorted by oldest first:

```bash
curl -H "Authorization: Bearer your_api_token" \
  "http://localhost:8080/api/cron-history?success=false&sort=asc&limit=5"
```

4. Get message job history with pagination and newest first sorting:

```bash
curl -H "Authorization: Bearer your_api_token" \
  "http://localhost:8080/api/cron-history?name=message&page=2&limit=15&sort=desc"
```

5. Get second page of all executions with 10 records per page:

```bash
curl -H "Authorization: Bearer your_api_token" \
  "http://localhost:8080/api/cron-history?page=2&limit=10"
```

6. Get history for a specific date range (from March 1st to March 15th, 2024):

```bash
curl -H "Authorization: Bearer your_api_token" \
  "http://localhost:8080/api/cron-history?start_date=2024-03-01&end_date=2024-03-15"
```

7. Get failed executions from the last week:

```bash
curl -H "Authorization: Bearer your_api_token" \
  "http://localhost:8080/api/cron-history?success=false&start_date=2024-03-08"
```

8. Get collect job history for a specific date with precise timestamps:

```bash
curl -H "Authorization: Bearer your_api_token" \
  "http://localhost:8080/api/cron-history?name=collect&start_date=2024-03-15T00:00:00Z&end_date=2024-03-15T23:59:59Z"
```

9. Get recent executions from the last 3 days, sorted oldest first:

```bash
curl -H "Authorization: Bearer your_api_token" \
  "http://localhost:8080/api/cron-history?start_date=2024-03-12&sort=asc"
```

**Date Range Filtering Notes:**

- **Supported date formats:**
  - Date only: `YYYY-MM-DD` (e.g., `2024-03-15`)
  - RFC3339 with timezone: `YYYY-MM-DDTHH:MM:SSZ` (e.g., `2024-03-15T10:30:00Z`)
- **Date validation:**
  - Invalid date formats will return a `400 Bad Request` error
  - If `start_date` is after `end_date`, the API will return a `400 Bad Request` error
- **Date range behavior:**
  - `start_date` is inclusive (records from this date onwards)
  - `end_date` is inclusive (records up to the end of this date)
  - When using date-only format, `end_date` includes the entire day (until 23:59:59.999...)
- **Timezone handling:**
  - All timestamps are stored and compared in UTC
  - When using date-only format, the date is interpreted as the start of the day in UTC

### Authentication

All API endpoints are protected with Bearer token authentication. You need to provide the `API_TOKEN` in the request header:

```bash
Authorization: Bearer your_api_token
```

### Available Endpoints

#### Get All Cron Settings

```http
GET /api/crons
```

Returns the current settings for all cron jobs.

Response example:

```json
[
  {
    "name": "collect",
    "schedule": "0 13 * * 6",
    "is_active": true,
    "updated_at": "2024-03-15T10:00:00Z"
  },
  {
    "name": "message",
    "schedule": "12 10 * * *",
    "is_active": true,
    "updated_at": "2024-03-15T10:00:00Z"
  }
]
```

#### Update Cron Schedule

```http
PUT /api/crons/{name}/schedule
```

Update the schedule for a specific cron job. The `name` can be either `collect` or `message`.

Request body:

```json
{
  "schedule": "0 15 * * 6"
}
```

Response example:

```json
{
  "status": "success",
  "message": "Schedule updated successfully"
}
```

#### Update Cron Status

```http
PUT /api/crons/{name}/status
```

Enable or disable a specific cron job. The `name` can be either `collect` or `message`.

Request body:

```json
{
  "is_active": false
}
```

Response example:

```json
{
  "status": "success",
  "message": "Status updated successfully"
}
```

#### Get Collect Settings

```http
GET /api/collect-settings
```

Returns the current settings for repository collection.

Response example:

```json
{
  "max_repos": 5,
  "since": "daily",
  "spoken_language_code": "en"
}
```

#### Update Collect Settings

```http
PUT /api/collect-settings
```

Update the repository collection settings.

Request body:

```json
{
  "max_repos": 10,
  "since": "weekly",
  "spoken_language_code": "uk"
}
```

Response example:

```json
{
  "status": "success",
  "message": "Collect settings updated successfully"
}
```

### Example Usage

1. Get all cron settings:

```bash
curl -H "Authorization: Bearer your_api_token" \
  http://localhost:8080/api/crons
```

2. Update collect schedule:

```bash
curl -X PUT \
  -H "Authorization: Bearer your_api_token" \
  -H "Content-Type: application/json" \
  -d '{"schedule": "0 15 * * 6"}' \
  http://localhost:8080/api/crons/collect/schedule
```

3. Disable message cron:

```bash
curl -X PUT \
  -H "Authorization: Bearer your_api_token" \
  -H "Content-Type: application/json" \
  -d '{"is_active": false}' \
  http://localhost:8080/api/crons/message/status
```

4. Get collect settings:

```bash
curl -H "Authorization: Bearer your_api_token" \
  http://localhost:8080/api/collect-settings
```

5. Update collect settings:

```bash
curl -X PUT \
  -H "Authorization: Bearer your_api_token" \
  -H "Content-Type: application/json" \
  -d '{
    "max_repos": 10,
    "since": "weekly",
    "spoken_language_code": "uk"
  }' \
  http://localhost:8080/api/collect-settings
```

### Data Persistence

All settings are stored in BadgerDB, a performant key-value store. The data is persisted in the `data/badger` directory. When running in Docker, make sure to configure a volume for this directory to maintain settings across container restarts.

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
