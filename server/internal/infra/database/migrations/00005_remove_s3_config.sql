-- +goose Up
-- Remove S3/GCS backup config rows; backups are now local snapshots
-- managed through the admin UI (/api/backup) or the CLI (divoene backup).
DELETE FROM config_settings WHERE key IN ('s3_endpoint', 's3_bucket', 's3_access_key', 's3_secret_key');

-- +goose Down
-- Do not restore S3 config; backup mechanism has changed to local snapshots.