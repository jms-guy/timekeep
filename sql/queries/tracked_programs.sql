-- name: GetProgramByName :one
SELECT * FROM tracked_programs
WHERE name = ?;

-- name: GetAllPrograms :many
SELECT * FROM tracked_programs;

-- name: AddProgram :one
INSERT INTO tracked_programs (name)
VALUES (?);

-- name: RemoveProgram :exec
DELETE FROM tracked_programs
WHERE name = ?;

-- name: UpdateLifetime :one
UPDATE tracked_programs
SET lifetime_seconds = lifetime_seconds + ?
WHERE id = ?;