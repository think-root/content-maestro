# [1.12.0](https://github.com/think-root/content-maestro/compare/v1.11.0...v1.12.0) (2025-05-27)


### Features

* add date range filtering to cron history API ([c8173b8](https://github.com/think-root/content-maestro/commit/c8173b84a7578ac0f58ca4e2752e0e8edf347bc9))

# [1.11.0](https://github.com/think-root/content-maestro/compare/v1.10.1...v1.11.0) (2025-05-27)


### Features

* add GetCronHistoryCount to improve cron history pagination ([3b9b027](https://github.com/think-root/content-maestro/commit/3b9b027e5d9244d9432f8c5d3b1017b688221438))
* add pagination metadata to cron history response ([f5e5787](https://github.com/think-root/content-maestro/commit/f5e5787e0f37e5d4b16d58ce41359ef0fca4c223))
* implement pagination and sorting for cron history API ([6d75855](https://github.com/think-root/content-maestro/commit/6d758552a3c123a9cf6ff3889e2dbda957321e1f))

## [1.10.1](https://github.com/think-root/content-maestro/compare/v1.10.0...v1.10.1) (2025-05-27)


### Bug Fixes

* allow nil success parameter in GetCronHistory for unfiltered results ([06257c0](https://github.com/think-root/content-maestro/commit/06257c0c833fb81b4f442b99912b0839389d2f4f))
* handle empty success parameter in GetCronHistory ([143765a](https://github.com/think-root/content-maestro/commit/143765ac6c28d48647bbb7bfb6fbbd42c665e6e1))

# [1.10.0](https://github.com/think-root/content-maestro/compare/v1.9.0...v1.10.0) (2025-05-27)


### Features

* add CronHistory model ([0875e7a](https://github.com/think-root/content-maestro/commit/0875e7a9f36b95c37902bfcc668ae12f134fc19f))
* add endpoint for retrieving cron history ([1a8c68f](https://github.com/think-root/content-maestro/commit/1a8c68f1a741c802c366d5b120cf0c45a1bdd62f))
* implement endpoint for retrieving paginated cron history ([34ee602](https://github.com/think-root/content-maestro/commit/34ee602120ff6827bf55d8f357bc3e8a1f68579b))
* implement LogCronExecution and GetCronHistory methods in store ([0b4ca21](https://github.com/think-root/content-maestro/commit/0b4ca21b55fc7f56af386b71141771a117abb863))
* log cron execution status for collect job ([cd41d08](https://github.com/think-root/content-maestro/commit/cd41d087a30e898d78d5bbd9f8a72bf25dff5c69))
* log cron execution status for message job ([29a8c18](https://github.com/think-root/content-maestro/commit/29a8c185a784fefe25500b679c2de83bcb15dc02))

# [1.9.0](https://github.com/think-root/content-maestro/compare/v1.8.1...v1.9.0) (2025-05-27)


### Features

* add LLM configuration to CollectJob for AI-powered descriptions ([ac64b9e](https://github.com/think-root/content-maestro/commit/ac64b9e211721608fbc06148a317c6001f821a3e))

## [1.8.1](https://github.com/think-root/content-maestro/compare/v1.8.0...v1.8.1) (2025-04-22)


### Bug Fixes

* improve socialify URL construction and request handling ([8e2f28b](https://github.com/think-root/content-maestro/commit/8e2f28b4c4588970a923e299e4dfd30ea140df15))

# [1.8.0](https://github.com/think-root/content-maestro/compare/v1.7.0...v1.8.0) (2025-03-31)


### Features

* add new endpoints for managing collect settings ([76313b5](https://github.com/think-root/content-maestro/commit/76313b5d75f06a7273f289e8f67345abc4f32c9e))
* **store:** add collect settings management ([20051e3](https://github.com/think-root/content-maestro/commit/20051e3013a07f1b97fdfbe6037feaa5aeae3eb1))

# [1.7.0](https://github.com/think-root/content-maestro/compare/v1.6.0...v1.7.0) (2025-03-28)


### Features

* add job initialization and scheduling functionality ([c458fa8](https://github.com/think-root/content-maestro/commit/c458fa8f7fdf385312a33cbfe39b5cd8329c565a))
* refactor main to dynamically initialize schedulers from Cron settings ([9bb2f2e](https://github.com/think-root/content-maestro/commit/9bb2f2e92d381b49332caa84b8907cc02bc82b4b))

# [1.6.0](https://github.com/think-root/content-maestro/compare/v1.5.7...v1.6.0) (2025-03-27)


### Features

* add JobFunc type and JobRegistry for scheduling jobs ([0a7fb8c](https://github.com/think-root/content-maestro/commit/0a7fb8cd4dba78b4aae39b738c781caa0442320c))
* implement API request handling in MessageJob and add InitJobs function ([52acf4e](https://github.com/think-root/content-maestro/commit/52acf4e8983a606c596458a145b21fbfbc7a0ee9))
* pass initialized jobs to CronAPI for enhanced job scheduling ([632dc16](https://github.com/think-root/content-maestro/commit/632dc164a5b7bdfa7633b97f495c7059ccfe4b96))
* refactor CronAPI to accept JobRegistry and streamline job scheduling ([f2d0313](https://github.com/think-root/content-maestro/commit/f2d03138cfa2a789d55b295d3fd3bcda49ab57bf))

## [1.5.7](https://github.com/think-root/content-maestro/compare/v1.5.6...v1.5.7) (2025-03-27)


### Bug Fixes

* replace utils function with os.MkdirAll for directory creation ([56cba04](https://github.com/think-root/content-maestro/commit/56cba041aa11d0d1071bdb45e68f0ec1e5f909c3))

## [1.5.6](https://github.com/think-root/content-maestro/compare/v1.5.5...v1.5.6) (2025-03-27)


### Bug Fixes

* update directory path for image storage to use relative path ([7e40948](https://github.com/think-root/content-maestro/commit/7e409482214fe7f2e2783c505476b54534a0537c))

## [1.5.5](https://github.com/think-root/content-maestro/compare/v1.5.4...v1.5.5) (2025-03-27)


### Bug Fixes

* change directory creation permissions to 0777 ([59a873a](https://github.com/think-root/content-maestro/commit/59a873a8fb265e8852d8d6dd6f33730207762256))

## [1.5.4](https://github.com/think-root/content-maestro/compare/v1.5.3...v1.5.4) (2025-03-27)


### Bug Fixes

* replace manual directory creation with os.MkdirAll for database path ([0f8bd0b](https://github.com/think-root/content-maestro/commit/0f8bd0b87013686a4050a7c9fcb8f9dafb31b523))

## [1.5.3](https://github.com/think-root/content-maestro/compare/v1.5.2...v1.5.3) (2025-03-27)


### Bug Fixes

* refactor directory creation to use utility function ([af986d7](https://github.com/think-root/content-maestro/commit/af986d789c30e0d064b1e8ca8367485f87f94a99))

## [1.5.2](https://github.com/think-root/content-maestro/compare/v1.5.1...v1.5.2) (2025-03-27)


### Bug Fixes

* add missing creation of tmp/gh_project_img dir needed for socialify image generation ([a48392b](https://github.com/think-root/content-maestro/commit/a48392bb38e96190d6ab5b478b493e087f585e86))

## [1.5.1](https://github.com/think-root/content-maestro/compare/v1.5.0...v1.5.1) (2025-03-27)


### Bug Fixes

* configure repository URL and bearer token using environment variables ([f3c4c47](https://github.com/think-root/content-maestro/commit/f3c4c47a0e665aab33928a016b3cfcc258dca313))

# [1.5.0](https://github.com/think-root/content-maestro/compare/v1.4.0...v1.5.0) (2025-03-27)


### Features

* add new store package for managing cron settings ([23b2e48](https://github.com/think-root/content-maestro/commit/23b2e48c12a5c56c53524279b68777b1da3e69b9))
* **cron:** add new cron jobs for messaging and collecting posts ([d27ea7b](https://github.com/think-root/content-maestro/commit/d27ea7bb3d8510a4887d8c0ffc90c1e8469f290b))
* **cron:** add new scheduler with cron job support ([4b45a75](https://github.com/think-root/content-maestro/commit/4b45a754e8bf761813440a3323d4589d375c0bc4))
* **middleware:** add logging middleware ([e324fa5](https://github.com/think-root/content-maestro/commit/e324fa5e8363f15266d1c7883fe09455101a48df))
* **scheduler:** add scheduler update function ([2ed98f1](https://github.com/think-root/content-maestro/commit/2ed98f13a343d89c49bda20106ce8665009d3878))

# [1.4.0](https://github.com/think-root/content-maestro/compare/v1.3.0...v1.4.0) (2025-03-27)


### Bug Fixes

* add delay to scheduler start ([b75d341](https://github.com/think-root/content-maestro/commit/b75d34101a9030134ace529e79201f34a7428d1c))


### Features

* **cron_api:** add cron expression validation ([c5b5b98](https://github.com/think-root/content-maestro/commit/c5b5b9875e4114500ce4e40e0616df946a70dd29))

# [1.3.0](https://github.com/think-root/content-maestro/compare/v1.2.0...v1.3.0) (2025-03-26)


### Features

* add CORS middleware for handling cross-origin requests ([1becdc7](https://github.com/think-root/content-maestro/commit/1becdc7dde22c909b7a37888257c48cb8cd59087))

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
