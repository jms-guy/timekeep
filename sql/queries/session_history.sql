-- name: AddToSessionHistory :exec
INSERT INTO session_history (program_name, start_time, end_time, duration_seconds)
VALUES (?, ?, ?, ?);

-- name: GetLastSessionForProgram :one 
SELECT * FROM session_history
WHERE session_history.program_name = ?
ORDER BY end_time DESC
LIMIT 1;

-- name: GetCountOfSessionsForProgram :one
SELECT COUNT(*) FROM session_history
WHERE session_history.program_name = ?;

-- name: GetAllSessionsForProgram :many
SELECT * FROM session_history
WHERE session_history.program_name = ?
ORDER BY start_time DESC;

-- name: RemoveAllRecords :exec
DELETE FROM session_history;

-- name: RemoveRecordsForProgram :exec
DELETE FROM session_history
WHERE session_history.program_name = ?;