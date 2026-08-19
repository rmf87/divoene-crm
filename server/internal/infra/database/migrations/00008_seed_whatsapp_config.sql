-- +goose Up
-- WhatsApp Business Cloud API settings, editable by the manager via the admin
-- Config panel. Empty values leave the client in mock mode; env fallback still
-- applies when the row is empty (see ConfigService.GetDecryptedValue).
INSERT OR IGNORE INTO config_settings (key, value, value_type, description) VALUES
('whatsapp_token', '', 'secret', 'WhatsApp Business Cloud API access token'),
('whatsapp_phone_number_id', '', 'string', 'WhatsApp phone number ID'),
('whatsapp_app_secret', '', 'secret', 'WhatsApp app secret (webhook signature)'),
('whatsapp_webhook_verify_token', '', 'secret', 'WhatsApp webhook verification token'),
('whatsapp_base_url', 'https://graph.facebook.com/v21.0', 'string', 'WhatsApp Graph API base URL');

-- +goose Down
DELETE FROM config_settings WHERE key LIKE 'whatsapp_%';
