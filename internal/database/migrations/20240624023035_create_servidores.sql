-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS servers (
    id SERIAL PRIMARY KEY,
    host VARCHAR(255),
    password VARCHAR(255),
    script TEXT,
    pipeline_id INTEGER,
    label VARCHAR(255),
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pipeline_id) REFERENCES pipelines (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS servidores;
-- +goose StatementEnd