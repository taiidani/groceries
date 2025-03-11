-- +goose Up
ALTER TABLE category ADD COLUMN description TEXT;

-- +goose Down
ALTER TABLE category DROP COLUMN description;
