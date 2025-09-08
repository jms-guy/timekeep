-- name: CreateSession :exec
INSERT INTO active_sessions (program_name, start_time)
VALUES (?, ?);