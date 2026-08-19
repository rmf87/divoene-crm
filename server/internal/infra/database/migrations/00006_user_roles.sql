-- +goose Up
-- Multi-role support: move the single users.role column into a
-- user_roles join table so each user can hold several roles.
CREATE TABLE user_roles (
    user_id TEXT NOT NULL,
    role TEXT NOT NULL,
    PRIMARY KEY (user_id, role),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Backfill each user's existing role.
INSERT OR IGNORE INTO user_roles (user_id, role)
SELECT id, role FROM users WHERE role != '';

-- The users.role column is no longer authoritative.
ALTER TABLE users DROP COLUMN role;

-- +goose Down
-- Restore the single-role model, keeping the lowest role of each user.
ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'seller';
UPDATE users SET role = (
    SELECT ur.role FROM user_roles ur WHERE ur.user_id = users.id ORDER BY ur.role LIMIT 1
);
DROP TABLE user_roles;
