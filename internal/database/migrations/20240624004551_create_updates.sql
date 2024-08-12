-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS updates (
    id SERIAL PRIMARY KEY,
    pusher_name TEXT,
    branch TEXT,
    status TEXT,
    message TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS updates;
-- +goose StatementEnd