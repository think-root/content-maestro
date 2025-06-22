# API Reference

The application provides a comprehensive REST API for managing cron jobs, repository collection settings, prompt settings, and cron job history.

## Authentication

All API endpoints are protected with Bearer token authentication. You need to provide the `API_TOKEN` in the request header:

```bash
Authorization: Bearer your_api_token
```

## Cron Job Management

### Get All Cron Settings

```http
GET /api/crons
```

Returns the current settings for all cron jobs.

**Response example:**

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

### Update Cron Schedule

```http
PUT /api/crons/{name}/schedule
```

Update the schedule for a specific cron job. The `name` can be either `collect` or `message`.

**Parameters:**
- `name` (path): Cron job name (`collect` or `message`)

**Request body:**

```json
{
  "schedule": "0 15 * * 6"
}
```

**Response example:**

```json
{
  "status": "success",
  "message": "Schedule updated successfully"
}
```

### Update Cron Status

```http
PUT /api/crons/{name}/status
```

Enable or disable a specific cron job. The `name` can be either `collect` or `message`.

**Parameters:**
- `name` (path): Cron job name (`collect` or `message`)

**Request body:**

```json
{
  "is_active": false
}
```

**Response example:**

```json
{
  "status": "success",
  "message": "Status updated successfully"
}
```

## Repository Collection Settings

### Get Collect Settings

```http
GET /api/collect-settings
```

Returns the current settings for repository collection.

**Response example:**

```json
{
  "max_repos": 5,
  "since": "daily",
  "spoken_language_code": "en"
}
```

### Update Collect Settings

```http
PUT /api/collect-settings
```

Update the repository collection settings.

**Request body:**

```json
{
  "max_repos": 10,
  "since": "weekly",
  "spoken_language_code": "uk"
}
```

**Parameters:**
- `max_repos` (integer): Maximum number of repositories to collect
- `since` (string): Time period for collection (`daily`, `weekly`, `monthly`)
- `spoken_language_code` (string): Language code for content (e.g., `en`, `uk`, `es`)

**Response example:**

```json
{
  "status": "success",
  "message": "Collect settings updated successfully"
}
```

## Prompt Settings

### Get Prompt Settings

```http
GET /api/prompt-settings
```

Returns the current AI prompt settings used for content generation.

**Response example:**

```json
{
  "use_direct_url": true,
  "llm_provider": "openrouter",
  "temperature": 0.1,
  "model": "openai/gpt-4o-mini-search-preview",
  "content": "You are a helpful AI assistant that generates engaging content about software repositories.",
  "updated_at": "2024-03-15T10:00:00Z"
}
```

### Update Prompt Settings

```http
POST /api/prompt-settings/update
```

Update the AI prompt settings for content generation.

**Request body:**

```json
{
  "use_direct_url": false,
  "llm_provider": "openrouter",
  "temperature": 0.1,
  "model": "openai/gpt-4o-mini",
  "content": "You are an expert technical writer specializing in open-source projects."
}
```

**Parameters:**
- `use_direct_url` (boolean, optional): Whether to use direct URL for LLM API calls
- `llm_provider` (string, optional): LLM provider name (e.g., `openai`, `mistral_agent`, `mistral_api`, `openrouter`)
- `temperature` (float, optional): Controls randomness in AI responses (0.0-2.0)
- `model` (string, optional): The AI model to use for content generation
- `content` (string, optional): The prompt content/template for AI generation

**Response example:**

```json
{
  "status": "success",
  "message": "Prompt settings updated successfully"
}
```

## Cron Job History

### Get Cron Job History

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

## Example Usage

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

3. Enable collect cron:

```bash
curl -X PUT \
  -H "Authorization: Bearer your_api_token" \
  -H "Content-Type: application/json" \
  -d '{"is_active": true}' \
  http://localhost:8080/api/crons/collect/status
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

6. Get prompt settings:

```bash
curl -H "Authorization: Bearer your_api_token" \
  http://localhost:8080/api/prompt-settings
```

7. Update prompt settings:

```bash
curl -X POST \
  -H "Authorization: Bearer your_api_token" \
  -H "Content-Type: application/json" \
  -d '{
    "use_direct_url": false,
    "llm_provider": "anthropic",
    "temperature": 0.8,
    "model": "anthropic/claude-3-sonnet",
    "content": "You are an expert technical writer specializing in open-source projects."
  }' \
  http://localhost:8080/api/prompt-settings/update
```

## Data Persistence

All settings are stored in PostgreSQL database. The data is persisted in the configured PostgreSQL instance. When running in Docker, make sure to configure the database connection properly using the environment variables specified in the configuration section.