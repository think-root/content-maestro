# [1.2.0](https://github.com/think-root/content-maestro/compare/v1.1.0...v1.2.0) (2025-03-26)


### Features

* add AuthMiddleware for API token validation in HTTP requests ([40a95e7](https://github.com/think-root/content-maestro/commit/40a95e7d7b035e5396c1f84aa00c9f1041dc37c2))
* add CronAPI for managing cron jobs with endpoints for retrieving, updating schedules, and statuses ([4aa06fe](https://github.com/think-root/content-maestro/commit/4aa06fef4aca483984a4eb2a305ecdcf7fbbe08e))
* add CronSetting and request types for cron job management ([37493f4](https://github.com/think-root/content-maestro/commit/37493f42a62a455c396c8d43aac6caf119868598))
* enhance CollectCron function to utilize store for dynamic scheduling ([a1e6448](https://github.com/think-root/content-maestro/commit/a1e6448400576e4597926c0f099a24341452c8e8))
* implement API for managing cron jobs ([0336a02](https://github.com/think-root/content-maestro/commit/0336a0245b0e2fca9cef67c84d0438ef9f196cea))
* implement Badger store for managing cron settings with CRUD operations ([69f38aa](https://github.com/think-root/content-maestro/commit/69f38aa3efadd16fb43a1641fdf19ab7bca1ab3e))
* modify MessageCron function to integrate store for dynamic cron scheduling ([914d374](https://github.com/think-root/content-maestro/commit/914d3740c617198be0f9f5183a4e76deccccdf31))

# [1.1.0](https://github.com/think-root/content-maestro/compare/v1.0.1...v1.1.0) (2025-03-26)


### Bug Fixes

* update repository retrieval to include sorting parameters ([0b247d9](https://github.com/think-root/content-maestro/commit/0b247d91c1c1ffcf435f2e1bd2c1d29e3e0b7ab7))


### Features

* enhance GetRepository function to support sorting parameters ([89d846c](https://github.com/think-root/content-maestro/commit/89d846c739dd7e90508076f907c364f7076fb7d4))

## [1.0.1](https://github.com/think-root/content-maestro/compare/v1.0.0...v1.0.1) (2025-03-07)


### Bug Fixes

* correct URL formatting in CollectCron request ([7cdd451](https://github.com/think-root/content-maestro/commit/7cdd451e51d684e3113b7eff4da4771e21473dbc))

# 1.0.0 (2025-03-04)


### Bug Fixes

* **api:** update environment variable names for Twitter and Telegram API configurations ([778f26a](https://github.com/think-root/content-maestro/commit/778f26ad1ff0a58ec9209338591d7f4721a8e152))


### Features

* **api:** add GetAPIConfigs function to retrieve API configuration ([18eef09](https://github.com/think-root/content-maestro/commit/18eef097e105fa49a4c965cac9d6d2200f865f69))
* **schedule:** add cron job to collect repositories ([14aac9f](https://github.com/think-root/content-maestro/commit/14aac9f3bbf8a3e04a1bf80375ab4c99dce4be83))
* **schedule:** add MessageCron function for scheduled repository updates ([a1ad802](https://github.com/think-root/content-maestro/commit/a1ad80239198f29ace3d8b8227af619866569705))
* **schedule:** enhance MessageCron to support dynamic API configurations and content types ([1075226](https://github.com/think-root/content-maestro/commit/10752269930df4bf5a4237da7aa54255e44c9766))
