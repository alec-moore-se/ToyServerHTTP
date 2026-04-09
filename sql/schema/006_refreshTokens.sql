-- +goose Up
CREATE TABLE refresh_tokens(
  token VARCHAR(32) PRIMARY KEY NOT NULL,
  created_at TIMESTAMP,
  updates_at TIMESTAMP,
  user_id UUID REFERENCES users(id)
ON DELETE CASCADE NOT NULL,
  expires_at TIMESTAMP,
  revoked_at TIMESTAMP
);
-- +goose Down
DROP TABLE refresh_tokens;
