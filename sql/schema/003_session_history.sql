-- +goose Up
CREATE TABLE session_history (
    id INTEGER PRIMARY KEY,
    program_name TEXT NOT NULL REFERENCES tracked_programs(name)
    ON DELETE CASCADE,
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    duration_seconds INTEGER NOT NULL
);

-- +goose Down
DROP TABLE session_history;