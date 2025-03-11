-- +goose Up
-- +goose StatementBegin
DELETE FROM item;
DELETE FROM category;

INSERT INTO category (id, name, description) VALUES
(1, 'Produce', 'Frozen foods'),
(2, 'Bulk Foods', 'Mostly nuts');

INSERT INTO item (id, category_id, name, quantity, done) VALUES
(1, 1, 'Breakfast sausage', '1 package', FALSE),
(2, 1, 'Tofu', '', FALSE),
(3, 1, 'Pizza', '2', TRUE),
(4, 2, 'Cashews', '1 cup', TRUE),
(5, 2, 'Garlic powder', '1.5oz', FALSE),
(6, 2, 'Almonds', '0.5lb', FALSE);
-- +goose StatementEnd

-- +goose Down
