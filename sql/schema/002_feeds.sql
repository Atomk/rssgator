-- +goose Up
CREATE TABLE feeds (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name TEXT NOT NULL,
    url TEXT UNIQUE NOT NULL,
    user_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE
);
COMMENT ON COLUMN feeds.name IS 'Name of the feed, e.g. "Hacker News RSS".';
COMMENT ON COLUMN feeds.url IS 'RSS endpoint/file URL.';
COMMENT ON COLUMN feeds.user_id IS 'The user that created the feed.';

-- +goose Down
DROP TABLE feeds;