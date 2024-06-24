-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "pipelines" (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "pipelines";
-- +goose StatementEnd
