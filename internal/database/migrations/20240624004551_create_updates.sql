-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "updates" (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			"pusher_name" TEXT,
			"branch" TEXT,
			"status" TEXT,
			"message" TEXT,
			"created_at" DATETIME DEFAULT CURRENT_TIMESTAMP,
			"updated_at" DATETIME DEFAULT CURRENT_TIMESTAMP
);
		
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "updates";
-- +goose StatementEnd
