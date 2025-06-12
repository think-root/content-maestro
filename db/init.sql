CREATE DATABASE "think-root-db" WITH OWNER = postgres ENCODING = 'UTF8';

\c "think-root-db"

CREATE TABLE IF NOT EXISTS maestro_cron_settings (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    schedule VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS maestro_cron_history (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN NOT NULL,
    output TEXT
);

CREATE TABLE IF NOT EXISTS maestro_collect_settings (
    id SERIAL PRIMARY KEY,
    max_repos INTEGER NOT NULL DEFAULT 5,
    since VARCHAR(50) NOT NULL DEFAULT 'daily',
    spoken_language_code VARCHAR(10) NOT NULL DEFAULT 'en'
);

INSERT INTO maestro_cron_settings (name, schedule, is_active, updated_at)
SELECT 'collect', '13 13 * * 6', FALSE, CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM maestro_cron_settings WHERE name = 'collect');

INSERT INTO maestro_cron_settings (name, schedule, is_active, updated_at)
SELECT 'message', '12 12 * * *', FALSE, CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM maestro_cron_settings WHERE name = 'message');

INSERT INTO maestro_collect_settings (max_repos, since, spoken_language_code)
SELECT 2, 'daily', 'en'
WHERE NOT EXISTS (SELECT 1 FROM maestro_collect_settings);
