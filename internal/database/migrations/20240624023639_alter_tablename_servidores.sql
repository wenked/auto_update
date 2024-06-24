-- +goose Up
-- +goose StatementBegin
ALTER TABLE servidores
  RENAME TO servers;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE servers
  RENAME TO servidores;
-- +goose StatementEnd
