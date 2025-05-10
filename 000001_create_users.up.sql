CREATE TABLE  IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username VARCHAR(255) UNIQUE NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  password bytea NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);


--How migrations work:
-- 1. Create a new file in the migrations directory: 
-- migrate create -seq -ext sql -dir ./migrations create_users    => creates up and down migrations
-- migrate -path=./migrations -database="postgres://postgres:mypassword@localhost/berta?sslmode=disable" up

