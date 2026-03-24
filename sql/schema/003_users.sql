-- +goose Up 
CREATE TABLE if Not exists users (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  email TEXT NOT NULL UNIQUE,
  hashed_password TEXT
);

CREATE TABLE if not exists chirps (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  body TEXT NOT NULL,
  user_id UUID references users(id) ON DELETE CASCADE
);

-- +goose Down 
DROP TABLE users;
DROP TABLE chirps;
