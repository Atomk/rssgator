-- name: CreatePost :one
INSERT INTO posts (id, title, created_at, updated_at, url, description, published_at, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetPostsForUser :many
-- Get all posts of feeds followed by the user.
WITH user_feeds AS (
    SELECT feed_id
    FROM feed_follows
    WHERE user_id = $1
)
SELECT posts.*
FROM user_feeds
INNER JOIN posts ON user_feeds.feed_id = posts.feed_id
ORDER BY posts.published_at DESC
LIMIT $2;
