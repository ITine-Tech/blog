CREATE TABLE IF NOT EXISTS posts(
  id BIGSERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  text TEXT NOT NULL,
  user_id UUID NOT NULL, -- UUID to match user table
  tags TEXT[] NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  CONSTRAINT fk_posts_users FOREIGN KEY (user_id) REFERENCES users(id) -- adds the relationship between the user_id here and the id of the User table
);