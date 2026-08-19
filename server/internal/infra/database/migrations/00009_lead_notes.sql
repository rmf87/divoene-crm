-- +goose Up
-- Lead notes (quick notes sellers add during the pipeline). Previously stored
-- only in the frontend localStorage — now server-side and auditable.
ALTER TABLE leads ADD COLUMN notes TEXT NOT NULL DEFAULT '[]';

-- +goose Down
ALTER TABLE leads DROP COLUMN notes;
