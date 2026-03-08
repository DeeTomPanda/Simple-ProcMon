-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_captured_at ON processes (captured_at);
CREATE INDEX idx_cpu_percent ON processes (cpu_percent);
CREATE INDEX idx_mem_rss ON processes (mem_rss);
-- +goose StatementEnd