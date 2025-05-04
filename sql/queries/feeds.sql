-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: GetNextFeedsToFetch :many
select * from feeds
order by last_fetch_at asc nulls first
limit $1;

-- name: MarkFeedAsFetched :one
update feeds
set last_fetch_at = Now(),
updated_at = Now()
where id = $1
returning *;



