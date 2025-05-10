CREATE DATABASE berta;

CREATE EXTENSION IF NOT EXISTS citext; -- this needs to be run if email should be in the citext format

-- I had to switch to the database and create the tables there with the +

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username VARCHAR(255) UNIQUE NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,  -- this can be a citext (then capitalization of letters doesn't make a difference - case insensitive)
  password TEXT NOT NULL,  --very bad: in plain text!!! Needs to be hashed! use: bytea NOT NULL (in bytes)
  created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

CREATE TABLE posts (
  id BIGSERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  text TEXT NOT NULL,
  user_id UUID NOT NULL, -- Use UUID to match user table
  tags TEXT[] NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  CONSTRAINT fk_posts_users FOREIGN KEY (user_id) REFERENCES users(id) -- adds the relationship between the user_id here and the id of the User
);

INSERT INTO users (email, username, password)
VALUES ('user1@example.com', 'user1', 'password123');

INSERT INTO posts (title, text, user_id, tags)
VALUES ('First Post', 'This is the first post', (SELECT id FROM users WHERE email = 'user1@example.com'), ARRAY['tech', 'politics']);

INSERT INTO comments (content, user_id)
VALUES ('This is the first comment'),
(SELECT user_id from users WHERE )