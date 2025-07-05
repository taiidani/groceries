-- +goose Up
-- +goose StatementBegin
INSERT INTO category (id, name, description) VALUES (0, 'Uncategorized', 'Default category for newly created items');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM category WHERE id = 0;
-- +goose StatementEnd
