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

-- name: RemoveAllRecords :exec
DELETE FROM session_history;

-- name: RemoveRecordsForProgram :exec
DELETE FROM session_history
WHERE session_history.program_name = ?;

-- name: GetSessionHistory :many
SELECT * FROM (
    SELECT * FROM session_history
    WHERE program_name = ?
    ORDER BY end_time DESC
    LIMIT ?
) AS results
ORDER BY end_time ASC;

-- name: GetSessionHistoryByDate :many
SELECT * FROM (
    SELECT * FROM session_history
    WHERE program_name = ? 
      AND start_time <= ? AND end_time >= ?
    ORDER BY start_time DESC
    LIMIT ?
) AS results
ORDER BY start_time ASC;

-- name: GetSessionHistoryByRange :many
SELECT * FROM (
    SELECT * FROM session_history
    WHERE program_name = ?
      AND start_time <= ? AND end_time >= ?
    ORDER BY start_time DESC
    LIMIT ?
) AS results
ORDER BY start_time ASC;

-- name: GetAllSessionHistory :many
SELECT * FROM (
    SELECT * FROM session_history
    ORDER BY end_time DESC
    LIMIT ?
) AS results
ORDER BY end_time ASC;

-- name: GetAllSessionHistoryByDate :many
SELECT * FROM (
    SELECT * FROM session_history
    WHERE start_time <= ? AND end_time >= ?
    ORDER BY start_time DESC
    LIMIT ?
) AS results
ORDER BY start_time ASC;

-- name: GetAllSessionHistoryByRange :many
SELECT * FROM (
    SELECT * FROM session_history
    WHERE start_time <= ? AND end_time >= ?
    ORDER BY start_time DESC
    LIMIT ?
) AS results
ORDER BY start_time ASC;