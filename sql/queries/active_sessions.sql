-- name: CreateActiveSession :exec
INSERT INTO active_sessions (program_name, start_time)
VALUES (?, ?);

-- name: GetActiveSession :one
SELECT start_time FROM active_sessions
WHERE program_name = ?;

-- name: RemoveActiveSession :exec
DELETE FROM active_sessions
WHERE program_name = ?;

-- name: RemoveAllSessions :exec
DELETE FROM active_sessions;