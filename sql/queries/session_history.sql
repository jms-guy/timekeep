-- name: AddToSessionHistory :exec
INSERT INTO session_history (program_name, start_time, end_time, duration_seconds)
VALUES (?, ?, ?, ?);