-- +goose Up 
CREATE TABLE tracked_programs (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    lifetime_seconds INTEGER DEFAULT 0
);

-- +goose Down
DROP TABLE tracked_programs;