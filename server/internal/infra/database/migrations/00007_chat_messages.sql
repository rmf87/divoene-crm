-- +goose Up
-- Chat messages between sellers/guides and leads via WhatsApp Business Cloud API.
CREATE TABLE chat_messages (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    lead_id       TEXT NOT NULL,
    wa_message_id TEXT DEFAULT '',
    direction     TEXT NOT NULL CHECK (direction IN ('seller','lead','system')),
    body          TEXT NOT NULL,
    status        TEXT DEFAULT 'sent',
    sent_at       TEXT NOT NULL DEFAULT (datetime('now')),
    created_at    TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (lead_id) REFERENCES leads(id) ON DELETE CASCADE
);

CREATE INDEX idx_chat_messages_lead ON chat_messages(lead_id, sent_at);
CREATE UNIQUE INDEX idx_chat_messages_wa ON chat_messages(wa_message_id)
    WHERE wa_message_id != '';

-- +goose Down
DROP INDEX idx_chat_messages_wa;
DROP INDEX idx_chat_messages_lead;
DROP TABLE chat_messages;
