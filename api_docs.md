# API

> [!IMPORTANT]
> All API requests must include an Authorization header in the following format:
> Authorization: Bearer <API_TOKEN>
>
> All endpoints return JSON responses with appropriate HTTP status codes

## Methods

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
  "resource": "github",
  "since": "daily",
  "spoken_language_code": "en",
  "period": "past_24_hours",
  "language": "All"
}
```

> [!NOTE]
> All fields are stored in the database, but only relevant fields are used based on the `resource` value:
>
> - **GitHub**: uses `since`, `spoken_language_code`
> - **OssInsight**: uses `period`, `language`

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
    "resource": "github",
    "since": "weekly",
    "spoken_language_code": "uk",
    "period": "past_24_hours",
    "language": "All"
  }' \
  http://localhost:8080/api/collect-settings
```

**Request Parameters:**

| Parameter              | Type    | Required | Description                                                                                           |
| ---------------------- | ------- | -------- | ----------------------------------------------------------------------------------------------------- |
| `max_repos`            | integer | No       | Maximum number of repositories to collect                                                             |
| `resource`             | string  | No       | Data source: `github` (default) or `ossinsight`                                                       |
| `since`                | string  | No       | **For GitHub**: Time period (`daily`, `weekly`, `monthly`)                                            |
| `spoken_language_code` | string  | No       | **For GitHub**: Spoken language filter (e.g., `en`, `uk`, `es`)                                       |
| `period`               | string  | No       | **For OssInsight**: Time period (`past_24_hours`, `past_week`, `past_month`, `past_3_months`)         |
| `language`             | string  | No       | **For OssInsight**: Programming language filter (e.g., `Python`, `All`)                               |

**Request Example:**

```json
{
  "max_repos": 10,
  "resource": "github",
  "since": "weekly",
  "spoken_language_code": "uk",
  "period": "past_24_hours",
  "language": "All"
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

| Parameter        | Type    | Required | Description                                                                        |
| ---------------- | ------- | -------- | ---------------------------------------------------------------------------------- |
| `use_direct_url` | boolean | No       | Whether to use direct URL for LLM API calls                                        |
| `llm_provider`   | string  | No       | LLM provider name (e.g., `openai`, `mistral_agent`, `mistral_api`, `openrouter`)   |
| `temperature`    | float   | No       | Controls randomness in AI responses (0.0-2.0)                                      |
| `model`          | string  | No       | The AI model to use for content generation                                         |
| `content`        | string  | No       | The prompt content/template for AI generation                                      |

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
| `status`     | integer | No       | Filter by execution status: `0` (Failure), `1` (Success), `2` (Partial)      |
| `start_date` | string  | No       | Filter records from this date onwards (format: `YYYY-MM-DD` or RFC3339)      |
| `end_date`   | string  | No       | Filter records up to this date (format: `YYYY-MM-DD` or RFC3339)             |

**Response Structure:**

The API returns a paginated response with the following structure.
Note that the `status` field uses integer status codes:

- `0`: Failure
- `1`: Success
- `2`: Partial Success

Structure:

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
      "status": 1,
      "output": "Successfully collected 5 repositories"
    },
    {
      "name": "message",
      "timestamp": "2024-03-15T10:05:00Z",
      "status": 0,
      "output": "Network error"
    },
    {
      "name": "message",
      "timestamp": "2024-03-15T10:10:00Z",
      "status": 2,
      "output": "Message sent to: telegram. Failed: bluesky"
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

1. Get history for specific job with custom pagination:

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  "http://localhost:8080/api/cron-history?name=collect&page=1&limit=10"
```

1. Get only failed executions sorted by oldest first:

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  "http://localhost:8080/api/cron-history?status=0&sort=asc&limit=5"
```

1. Get message job history with pagination and newest first sorting:

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  "http://localhost:8080/api/cron-history?name=message&page=2&limit=15&sort=desc"
```

1. Get second page of all executions with 10 records per page:

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  "http://localhost:8080/api/cron-history?page=2&limit=10"
```

1. Get history for a specific date range (from March 1st to March 15th, 2024):

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  "http://localhost:8080/api/cron-history?start_date=2024-03-01&end_date=2024-03-15"
```

1. Get failed executions from the last week:

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  "http://localhost:8080/api/cron-history?status=0&start_date=2024-03-08"
```

1. Get collect job history for a specific date with precise timestamps:

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  "http://localhost:8080/api/cron-history?name=collect&start_date=2024-03-15T00:00:00Z&end_date=2024-03-15T23:59:59Z"
```

1. Get recent executions from the last 3 days, sorted oldest first:

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

## API Configurations Management

### /api/api-configs

**Endpoint:** `/api/api-configs`

**Method:** `GET`

**Description:** Retrieve all API configurations for external integrations (Twitter, Telegram, Bluesky, etc.).

**Curl Example:**

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  http://localhost:8080/api/api-configs
```

**Response Example:**

```json
[
  {
    "id": 1,
    "name": "twitter",
    "url": "{env.TWITTER_URL}/x/api/posts/create",
    "method": "POST",
    "auth_type": "api_key",
    "token_env_var": "TWITTER_API_KEY",
    "token_header": "X-API-Key",
    "content_type": "multipart",
    "timeout": 30,
    "success_code": 200,
    "enabled": true,
    "response_type": "json",
    "text_language": "en",
    "socialify_image": false,
    "default_json_body": "",
    "updated_at": "2024-03-15T10:00:00Z"
  },
  {
    "id": 2,
    "name": "telegram",
    "url": "{env.TELEGRAM_SERVER_URL}/telegram/send-message",
    "method": "POST",
    "auth_type": "api_key",
    "token_env_var": "TELEGRAM_SERVER_TOKEN",
    "token_header": "X-API-Key",
    "content_type": "multipart",
    "timeout": 30,
    "success_code": 200,
    "enabled": true,
    "response_type": "json",
    "text_language": "uk",
    "socialify_image": true,
    "default_json_body": "",
    "updated_at": "2024-03-15T10:00:00Z"
  }
]
```

### /api/api-configs/{name}

**Endpoint:** `/api/api-configs/{name}`

**Method:** `GET`

**Description:** Retrieve a specific API configuration by name.

**Curl Example:**

```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  http://localhost:8080/api/api-configs/twitter
```

**Response Example:**

```json
{
  "id": 1,
  "name": "twitter",
  "url": "{env.TWITTER_URL}/x/api/posts/create",
  "method": "POST",
  "auth_type": "api_key",
  "token_env_var": "TWITTER_API_KEY",
  "token_header": "X-API-Key",
  "content_type": "multipart",
  "timeout": 30,
  "success_code": 200,
  "enabled": true,
  "response_type": "json",
  "text_language": "en",
  "socialify_image": false,
  "default_json_body": "",
  "updated_at": "2024-03-15T10:00:00Z"
}
```

### /api/api-configs (create)

**Endpoint:** `/api/api-configs`

**Method:** `POST`

**Description:** Create a new API configuration for external integration.

**Curl Example:**

```bash
curl -X POST \
  -H "Authorization: Bearer <API_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "bluesky",
    "url": "{env.BLUESKY_URL}/bluesky/api/posts/create",
    "method": "POST",
    "auth_type": "api_key",
    "token_env_var": "BLUESKY_SERVER_KEY",
    "token_header": "X-API-Key",
    "content_type": "multipart",
    "timeout": 30,
    "success_code": 200,
    "enabled": true,
    "response_type": "json",
    "text_language": "en",
    "socialify_image": false,
    "default_json_body": ""
  }' \
  http://localhost:8080/api/api-configs
```

**Request Parameters:**

| Parameter           | Type    | Required | Description                                                                      |
| ------------------- | ------- | -------- | -------------------------------------------------------------------------------- |
| `name`              | string  | Yes      | Unique identifier for the API (alphanumeric, hyphens, underscores only)          |
| `url`               | string  | Yes      | The endpoint URL (supports `{env.VAR}` syntax)                                   |
| `method`            | string  | Yes      | HTTP method (GET, POST, PUT, DELETE, PATCH)                                      |
| `auth_type`         | string  | No       | Authentication type: `bearer`, `api_key`, or empty                               |
| `token_env_var`     | string  | No       | Environment variable name containing the auth token                              |
| `token_header`      | string  | No       | Header name for API key (required if `auth_type` is `api_key`)                   |
| `content_type`      | string  | Yes      | Request content type: `json` or `multipart`                                      |
| `timeout`           | integer | Yes      | Request timeout in seconds (must be > 0)                                         |
| `success_code`      | integer | Yes      | Expected HTTP success code (100-599)                                             |
| `enabled`           | boolean | Yes      | Whether the API is enabled                                                       |
| `response_type`     | string  | No       | Expected response format                                                         |
| `text_language`     | string  | No       | Language code for text content (e.g., `en`, `uk`)                                |
| `socialify_image`   | boolean | Yes      | Whether to generate socialify images                                             |
| `default_json_body` | string  | No       | JSON string of default key/value pairs (supports `{env.VAR}`)                    |

**Request Example:**

```json
{
  "name": "bluesky",
  "url": "{env.BLUESKY_URL}/bluesky/api/posts/create",
  "method": "POST",
  "auth_type": "api_key",
  "token_env_var": "BLUESKY_SERVER_KEY",
  "token_header": "X-API-Key",
  "content_type": "multipart",
  "timeout": 30,
  "success_code": 200,
  "enabled": true,
  "response_type": "json",
  "text_language": "en",
  "socialify_image": false,
  "default_json_body": ""
}
```

**Response Example:**

```json
{
  "id": 3,
  "name": "bluesky",
  "url": "{env.BLUESKY_URL}/bluesky/api/posts/create",
  "method": "POST",
  "auth_type": "api_key",
  "token_env_var": "BLUESKY_SERVER_KEY",
  "token_header": "X-API-Key",
  "content_type": "multipart",
  "timeout": 30,
  "success_code": 200,
  "enabled": true,
  "response_type": "json",
  "text_language": "en",
  "socialify_image": false,
  "default_json_body": "",
  "updated_at": "2024-03-15T10:00:00Z"
}
```

### /api/api-configs/{name} (update)

**Endpoint:** `/api/api-configs/{name}`

**Method:** `PUT`

**Description:** Update an existing API configuration. All fields are optional - only provide the fields you want to update.

**Curl Example:**

```bash
curl -X PUT \
  -H "Authorization: Bearer <API_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": false,
    "timeout": 60
  }' \
  http://localhost:8080/api/api-configs/twitter
```

**Request Parameters:**

All fields are optional. Only include the fields you want to update.

| Parameter           | Type    | Description                                                                      |
| ------------------- | ------- | -------------------------------------------------------------------------------- |
| `url`               | string  | The endpoint URL (supports `{env.VAR}` syntax)                                   |
| `method`            | string  | HTTP method (GET, POST, PUT, DELETE, PATCH)                                      |
| `auth_type`         | string  | Authentication type: `bearer`, `api_key`, or empty                               |
| `token_env_var`     | string  | Environment variable name containing the auth token                              |
| `token_header`      | string  | Header name for API key                                                          |
| `content_type`      | string  | Request content type: `json` or `multipart`                                      |
| `timeout`           | integer | Request timeout in seconds (must be > 0)                                         |
| `success_code`      | integer | Expected HTTP success code (100-599)                                             |
| `enabled`           | boolean | Whether the API is enabled                                                       |
| `response_type`     | string  | Expected response format                                                         |
| `text_language`     | string  | Language code for text content (e.g., `en`, `uk`)                                |
| `socialify_image`   | boolean | Whether to generate socialify images                                             |
| `default_json_body` | string  | JSON string of default key/value pairs (supports `{env.VAR}`)                    |

**Request Example:**

```json
{
  "enabled": false,
  "timeout": 60,
  "text_language": "es"
}
```

**Response Example:**

```json
{
  "id": 1,
  "name": "twitter",
  "url": "{env.TWITTER_URL}/x/api/posts/create",
  "method": "POST",
  "auth_type": "api_key",
  "token_env_var": "TWITTER_API_KEY",
  "token_header": "X-API-Key",
  "content_type": "multipart",
  "timeout": 60,
  "success_code": 200,
  "enabled": false,
  "response_type": "json",
  "text_language": "es",
  "socialify_image": false,
  "default_json_body": "",
  "updated_at": "2024-03-15T11:00:00Z"
}
```

### /api/api-configs/{name} (delete)

**Endpoint:** `/api/api-configs/{name}`

**Method:** `DELETE`

**Description:** Delete an API configuration.

**Curl Example:**

```bash
curl -X DELETE \
  -H "Authorization: Bearer <API_TOKEN>" \
  http://localhost:8080/api/api-configs/twitter
```

**Response Example:**

```json
{
  "status": "success",
  "message": "API config deleted successfully"
}
```

**API Configuration Notes:**

- **Environment Variables:** Use `{env.VARIABLE_NAME}` syntax in `url` and `default_json_body` fields to reference environment variables
- **Default JSON Body:** For APIs with `content_type: json`, you can specify default key/value pairs that are always included in requests. Store as a JSON string, e.g., `{"type": "chat", "jid": "{env.WAPP_JID}"}`
- **Auto-Reload:** After creating, updating, or deleting an API configuration, the system automatically reloads all configurations to apply changes immediately
- **Migration:** On first startup (v3.4.0+), existing configurations from `apis-config.yml` are automatically migrated to the database

**Status Codes:**

- 200: Success (GET, PUT)
- 201: Created (POST)
- 400: Bad Request - Invalid parameters or validation errors
- 401: Unauthorized - Invalid or missing Bearer token
- 404: Not Found - API configuration does not exist (GET, PUT, DELETE)
- 500: Internal Server Error - Database or server error
