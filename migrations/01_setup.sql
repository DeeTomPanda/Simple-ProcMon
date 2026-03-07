-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS processes (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    pid         INTEGER NOT NULL,
    name        TEXT NOT NULL,
    cpu_percent REAL,
    mem_rss     INTEGER,
    status      TEXT,
    captured_at DATETIME NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS process_events (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    pid         INTEGER NOT NULL,
    name        TEXT NOT NULL,
    event_type  TEXT NOT NULL,
    occurred_at DATETIME NOT NULL
);
-- +goose StatementEnd