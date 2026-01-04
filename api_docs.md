# API

> [!IMPORTANT]
> All API requests must include an Authorization header in the following format:
> Authorization: Bearer <API_TOKEN>
>
> All endpoints return JSON responses with appropriate HTTP status codes

### /api/crons

**Endpoint:** `/api/crons`

**Method:** `GET`

**Description:** Returns the current settings for all cron jobs.

**Curl Example:**

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  http://localhost:8080/api/crons
```

**Response Example:**

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

### /api/crons/{name}/schedule

**Endpoint:** `/api/crons/{name}/schedule`

**Method:** `PUT`

**Description:** Update the schedule for a specific cron job. The `name` can be either `collect` or `message`.

**Curl Example:**

```bash
curl -X PUT \
  -H "Authorization: Bearer <API_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"schedule": "0 15 * * 6"}' \
  http://localhost:8080/api/crons/collect/schedule
```

**Request Parameters:**

| Parameter  | Type   | Required | Description                                   |
| ---------- | ------ | -------- | --------------------------------------------- |
| `name`     | string | Yes      | Cron job name (`collect` or `message`)        |
| `schedule` | string | Yes      | Cron schedule expression (e.g., `0 15 * * 6`) |

**Request Example:**

```json
{
  "schedule": "0 15 * * 6"
}
```

**Response Example:**

```json
{
  "status": "success",
  "message": "Schedule updated successfully"
}
```

### /api/crons/{name}/status

**Endpoint:** `/api/crons/{name}/status`

**Method:** `PUT`

**Description:** Enable or disable a specific cron job. The `name` can be either `collect` or `message`.

**Curl Example:**

```bash
curl -X PUT \
  -H "Authorization: Bearer <API_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"is_active": true}' \
  http://localhost:8080/api/crons/collect/status
```

**Request Parameters:**

| Parameter   | Type    | Required | Description                              |
| ----------- | ------- | -------- | ---------------------------------------- |
| `name`      | string  | Yes      | Cron job name (`collect` or `message`)   |
| `is_active` | boolean | Yes      | Enable (true) or disable (false) the job |

**Request Example:**

```json
{
  "is_active": false
}
```

**Response Example:**

```json
{
  "status": "success",
  "message": "Status updated successfully"
}
```

### /api/collect-settings

**Endpoint:** `/api/collect-settings`

**Method:** `GET`

**Description:** Returns the current settings for repository collection.

**Curl Example:**

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  http://localhost:8080/api/collect-settings
```

**Response Example:**

```json
{
  "max_repos": 5,
  "since": "daily",
  "spoken_language_code": "en"
}
```

### /api/collect-settings (update)

**Endpoint:** `/api/collect-settings`

**Method:** `PUT`

**Description:** Update the repository collection settings.

**Curl Example:**

```bash
curl -X PUT \
  -H "Authorization: Bearer <API_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "max_repos": 10,
    "since": "weekly",
    "spoken_language_code": "uk"
  }' \
  http://localhost:8080/api/collect-settings
```

**Request Parameters:**

| Parameter              | Type    | Required | Description                                                        |
| ---------------------- | ------- | -------- | ------------------------------------------------------------------ |
| `max_repos`            | integer | No       | Maximum number of repositories to collect                          |
| `since`                | string  | No       | Time period for collection (`daily`, `weekly`, `monthly`)          |
| `spoken_language_code` | string  | No       | Language code for content (e.g., `en`, `uk`, `es`)                 |

**Request Example:**

```json
{
  "max_repos": 10,
  "since": "weekly",
  "spoken_language_code": "uk"
}
```

**Response Example:**

```json
{
  "status": "success",
  "message": "Collect settings updated successfully"
}
```

### /api/prompt-settings

**Endpoint:** `/api/prompt-settings`

**Method:** `GET`

**Description:** Returns the current AI prompt settings used for content generation.

**Curl Example:**

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  http://localhost:8080/api/prompt-settings
```

**Response Example:**

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

### /api/prompt-settings (update)

**Endpoint:** `/api/prompt-settings`

**Method:** `PUT`

**Description:** Update the AI prompt settings for content generation.

**Curl Example:**

```bash
curl -X PUT \
  -H "Authorization: Bearer <API_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "use_direct_url": false,
    "llm_provider": "anthropic",
    "temperature": 0.8,
    "model": "anthropic/claude-3-sonnet",
    "content": "You are an expert technical writer specializing in open-source projects."
  }' \
  http://localhost:8080/api/prompt-settings
```

**Request Parameters:**

| Parameter        | Type    | Required | Description                                                                              |
| ---------------- | ------- | -------- | ---------------------------------------------------------------------------------------- |
| `use_direct_url` | boolean | No       | Whether to use direct URL for LLM API calls                                              |
| `llm_provider`   | string  | No       | LLM provider name (e.g., `openai`, `mistral_agent`, `mistral_api`, `openrouter`)        |
| `temperature`    | float   | No       | Controls randomness in AI responses (0.0-2.0)                                            |
| `model`          | string  | No       | The AI model to use for content generation                                               |
| `content`        | string  | No       | The prompt content/template for AI generation                                            |

**Request Example:**

```json
{
  "use_direct_url": false,
  "llm_provider": "openrouter",
  "temperature": 0.1,
  "model": "openai/gpt-4o-mini",
  "content": "You are an expert technical writer specializing in open-source projects."
}
```

**Response Example:**

```json
{
  "status": "success",
  "message": "Prompt settings updated successfully"
}
```

### /api/cron-history

**Endpoint:** `/api/cron-history`

**Method:** `GET`

**Description:** Retrieve the history of cron job executions with pagination, sorting, and filtering.

**Curl Example:**

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  "http://localhost:8080/api/cron-history?name=collect&page=1&limit=10"
```

**Request Parameters:**

| Parameter    | Type    | Required | Description                                                                  |
| ------------ | ------- | -------- | ---------------------------------------------------------------------------- |
| `name`       | string  | No       | Filter by cron job name (`collect` or `message`)                             |
| `page`       | integer | No       | Page number (default: 1)                                                     |
| `limit`      | integer | No       | Number of records per page (default: 20)                                     |
| `sort`       | string  | No       | Sort order by execution date (`asc` or `desc`, default: `desc`)              |
| `success`    | boolean | No       | Filter by execution status (`true` or `false`)                               |
| `start_date` | string  | No       | Filter records from this date onwards (format: `YYYY-MM-DD` or RFC3339)      |
| `end_date`   | string  | No       | Filter records up to this date (format: `YYYY-MM-DD` or RFC3339)             |

**Response Structure:**

The API returns a paginated response with the following structure:

- `data`: Array of cron history records
- `pagination`: Pagination metadata object containing:
  - `total_count`: Total number of records matching the filters
  - `current_page`: Current page number
  - `total_pages`: Total number of pages available
  - `has_next`: Boolean indicating if there's a next page
  - `has_previous`: Boolean indicating if there's a previous page

**Response Example:**

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

**Usage Examples:**

1. Get all history (default: newest first, 20 records per page):

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  http://localhost:8080/api/cron-history
```

2. Get history for specific job with custom pagination:

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  "http://localhost:8080/api/cron-history?name=collect&page=1&limit=10"
```

3. Get only failed executions sorted by oldest first:

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  "http://localhost:8080/api/cron-history?success=false&sort=asc&limit=5"
```

4. Get message job history with pagination and newest first sorting:

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  "http://localhost:8080/api/cron-history?name=message&page=2&limit=15&sort=desc"
```

5. Get second page of all executions with 10 records per page:

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  "http://localhost:8080/api/cron-history?page=2&limit=10"
```

6. Get history for a specific date range (from March 1st to March 15th, 2024):

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  "http://localhost:8080/api/cron-history?start_date=2024-03-01&end_date=2024-03-15"
```

7. Get failed executions from the last week:

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  "http://localhost:8080/api/cron-history?success=false&start_date=2024-03-08"
```

8. Get collect job history for a specific date with precise timestamps:

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  "http://localhost:8080/api/cron-history?name=collect&start_date=2024-03-15T00:00:00Z&end_date=2024-03-15T23:59:59Z"
```

9. Get recent executions from the last 3 days, sorted oldest first:

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
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

**Status Codes:**

- 200: Success
- 400: Bad Request - Invalid parameters or date validation errors
- 401: Unauthorized - Invalid or missing Bearer token
- 500: Internal Server Error - Database or server error

