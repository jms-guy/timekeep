-- +goose Up
CREATE TABLE active_sessions (
    id INTEGER PRIMARY KEY,
    program_name TEXT UNIQUE NOT NULL REFERENCES tracked_programs(name)
    ON DELETE CASCADE,
    start_time DATETIME NOT NULL
);

-- +goose Down
DROP TABLE active_sessions;