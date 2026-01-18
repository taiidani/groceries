-- +goose Up
-- +goose StatementBegin

-- Repeat the row added via the migrations
INSERT INTO category (id, name, store_id, description) VALUES (0, 'Uncategorized', 0, 'Default category for newly created items');

INSERT INTO category (name, store_id, description) VALUES
('Produce', 1, 'Only the freshest'),
('Bulk Foods', 2, 'Mostly nuts'),
('Exotic Pets', 2, 'Not a frequented aisle'),
('Household Items', 1, ''),
('Empty Void', 0, 'Bereft of items');
-- +goose StatementEnd

-- +goose Down
