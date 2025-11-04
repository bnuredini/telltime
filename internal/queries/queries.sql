-- name: GetEvent :one
SELECT *
FROM event
WHERE id = ?;

-- name: GetEvents :many
SELECT start_time, window_class, window_title, duration
FROM event
ORDER BY start_time DESC;

-- name: GetEventsByTime :many
SELECT start_time, window_class, window_title, duration
FROM event
WHERE start_time BETWEEN sqlc.arg(start_time) AND sqlc.arg(end_time)
ORDER BY start_time DESC;

-- name: InsertEvents :exec
INSERT INTO event (start_time, window_class, window_title, duration)
VALUES (?, ?, ?, ?);
