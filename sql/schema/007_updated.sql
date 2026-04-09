-- +goose Up
ALTER TABLE refresh_tokens RENAME COLUMN updates_at TO updated_at;
-- +goose Down
ALTER TABLE refresh_tokens RENAME COLUMN updated_at TO updates_at;
