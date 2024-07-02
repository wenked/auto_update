-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS notification_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT,
		name TEXT,
		number TEXT,
		user_id INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
	);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notification_config;
-- +goose StatementEnd
