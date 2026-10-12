-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name VARCHAR(20) UNIQUE NOT NULL
);
COMMENT ON COLUMN users.name IS 'The username.';

-- +goose Down
DROP TABLE users;