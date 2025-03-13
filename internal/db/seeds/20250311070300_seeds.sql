-- +goose Up
-- +goose StatementBegin
DELETE FROM item;
ALTER SEQUENCE item_id_seq RESTART WITH 1;
DELETE FROM category;
ALTER SEQUENCE category_id_seq RESTART WITH 1;

INSERT INTO category (name, description) VALUES
('Produce', 'Frozen foods'),
('Bulk Foods', 'Mostly nuts'),
('Exotic Pets', 'Not a frequented aisle');

INSERT INTO item (category_id, name, quantity, done) VALUES
(1, 'Breakfast sausage', '1 package', FALSE),
(1, 'Tofu', '', FALSE),
(1, 'Pizza', '2', TRUE),
(2, 'Cashews', '1 cup', TRUE),
(2, 'Garlic powder', '1.5oz', FALSE),
(2, 'Almonds', '0.5lb', FALSE);
-- +goose StatementEnd

-- +goose Down
