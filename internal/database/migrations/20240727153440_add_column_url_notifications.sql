-- +goose Up
-- +goose StatementBegin
ALTER TABLE notification_config ADD COLUMN url TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE notification_config DROP COLUMN url;
-- +goose StatementEnd