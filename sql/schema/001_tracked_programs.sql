-- +goose Up 
CREATE TABLE tracked_programs (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    category TEXT,
    lifetime_seconds INTEGER NOT NULL DEFAULT 0
);

-- +goose Down
DROP TABLE tracked_programs;