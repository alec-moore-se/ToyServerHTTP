-- +goose Up 
CREATE TABLE users (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  email TEXT NOT NULL UNIQUE,
  hashed_password TEXT
);

-- +goose Down 
DROP TABLE users;
