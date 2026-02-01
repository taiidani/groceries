-- +goose Up
-- +goose StatementBegin
ALTER TABLE category DROP CONSTRAINT category_name_key;
ALTER TABLE category ADD CONSTRAINT category_name_store_unique UNIQUE (name, store_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE category DROP CONSTRAINT category_name_store_unique;
ALTER TABLE category ADD CONSTRAINT category_name_key UNIQUE (name);
-- +goose StatementEnd
