CREATE TABLE IF NOT EXISTS user_invitations (
    token bytea PRIMARY KEY,
    id UUID DEFAULT gen_random_uuid()
)