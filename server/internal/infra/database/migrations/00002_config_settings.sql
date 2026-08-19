-- +goose Up
CREATE TABLE config_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT '',
    value_type TEXT NOT NULL DEFAULT 'string',
    description TEXT NOT NULL DEFAULT '',
    updated_at TEXT DEFAULT (datetime('now')),
    updated_by TEXT DEFAULT ''
);

INSERT OR IGNORE INTO config_settings (key, value, value_type, description) VALUES
('clicksign_api_key', '', 'secret', 'Clicksign API access token'),
('clicksign_base_url', 'https://sandbox.clicksign.com/api/v3', 'string', 'Clicksign base URL'),
('clicksign_webhook_secret', '', 'secret', 'Clicksign webhook HMAC secret'),
('clicksign_template_key', '', 'string', 'Clicksign document template key'),
('openpix_app_id', '', 'secret', 'OpenPix/Woovi App ID'),
('openpix_base_url', 'https://api.woovi-sandbox.com/api', 'string', 'OpenPix base URL');

-- +goose Down
DROP TABLE IF EXISTS config_settings;
