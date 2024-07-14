-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys = OFF;
CREATE TABLE IF NOT EXISTS new_users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    email TEXT,
    password TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    company_id INTEGER,
    CONSTRAINT fk_company FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE SET NULL);


INSERT INTO new_users (name, email, password, created_at, updated_at) SELECT name, email, password, created_at, updated_at FROM users;


DROP TABLE users;

ALTER TABLE new_users RENAME TO users;
PRAGMA foreign_keys = ON;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN company_id;

-- +goose StatementEnd
