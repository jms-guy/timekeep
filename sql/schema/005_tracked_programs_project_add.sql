-- +goose Up
ALTER TABLE tracked_programs
ADD project TEXT;

-- +goose Down
ALTER TABLE tracked_programs
DROP COLUMN project;