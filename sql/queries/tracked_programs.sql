-- name: GetProgramByName :one
SELECT * FROM tracked_programs
WHERE name = ?;

-- name: GetAllProgramNames :many
SELECT name FROM tracked_programs;

-- name: GetAllPrograms :many
SELECT * FROM tracked_programs;

-- name: AddProgram :exec
INSERT OR IGNORE INTO tracked_programs (name)
VALUES (?);

-- name: RemoveProgram :exec
DELETE FROM tracked_programs
WHERE name = ?;

-- name: UpdateLifetime :exec
UPDATE tracked_programs
SET lifetime_seconds = lifetime_seconds + ?
WHERE name = ?;

-- name: RemoveAllPrograms :exec
DELETE FROM tracked_programs;