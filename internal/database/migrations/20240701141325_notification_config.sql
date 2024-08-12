-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS notification_config (
    id SERIAL PRIMARY KEY,
    type VARCHAR(255),
    name VARCHAR(255),
    number VARCHAR(255),
    user_id INTEGER,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    url VARCHAR(255),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notification_config;
-- +goose StatementEnd