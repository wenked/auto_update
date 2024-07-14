-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS new_pipelines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    "user_id" INTEGER,
    FOREIGN KEY (user_id) REFERENCES users(id)
);


INSERT INTO new_pipelines (id, name, created_at, updated_at)
SELECT id, name, created_at, updated_at
FROM pipelines;

DROP TABLE pipelines;

ALTER TABLE new_pipelines RENAME TO pipelines;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE pipelines DROP COLUMN user_id;

-- +goose StatementEnd
