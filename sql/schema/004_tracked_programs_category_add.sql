-- +goose Up
ALTER TABLE tracked_programs
ADD category TEXT;

-- +goose Down
ALTER TABLE tracked_programs
DROP COLUMN category;