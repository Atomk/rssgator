-- +goose Up
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    url TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL,
    published_at TIMESTAMP,
    feed_id UUID NOT NULL
        REFERENCES feeds(id)
        ON DELETE CASCADE
);
COMMENT ON COLUMN posts.url IS 'The URL of the post.';
COMMENT ON COLUMN posts.description IS 'A description of the post.';
COMMENT ON COLUMN posts.published_at IS 'The time the post was published. Can be NULL in case of an unsupported datetime format.';
COMMENT ON COLUMN posts.feed_id IS 'The ID of the feed that the post came from.';

-- +goose Down
DROP TABLE posts;