-- +goose Up 
ALTER TABLE users 
  ALTER COLUMN id SET NOT NULL,
  ALTER COLUMN hashed_password SET NOT NULL
  ;

ALTER TABLE chirps 
  ALTER COLUMN id SET NOT NULL
  ;
  
-- +goose Down 
ALTER TABLE users 
  ALTER COLUMN id DROP NOT NULL,
  ALTER COLUMN hashed_password DROP NOT NULL
  ;

ALTER TABLE chirps 
  ALTER COLUMN id DROP NOT NULL
  ;

