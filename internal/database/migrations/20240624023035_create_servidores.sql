-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS servidores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		host TEXT,
		password TEXT,
		script TEXT,
		pipeline_id,
		label TEXT,
		active BOOLEAN DEFAULT TRUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (pipeline_id) REFERENCES pipelines (id) ON DELETE CASCADE
	);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS servidores;
-- +goose StatementEnd
