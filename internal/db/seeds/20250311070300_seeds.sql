-- +goose Up
-- +goose StatementBegin
DELETE FROM item_bag;
ALTER SEQUENCE item_bag_id_seq RESTART WITH 1;
DELETE FROM item_list;
ALTER SEQUENCE item_list_id_seq RESTART WITH 1;
DELETE FROM item;
ALTER SEQUENCE item_id_seq RESTART WITH 1;
DELETE FROM category;
ALTER SEQUENCE category_id_seq RESTART WITH 1;

-- Repeat the row added via the migrations
INSERT INTO category (id, name, description) VALUES (0, 'Uncategorized', 'Default category for newly created items');

INSERT INTO category (name, description) VALUES
('Produce', 'Frozen foods'),
('Bulk Foods', 'Mostly nuts'),
('Exotic Pets', 'Not a frequented aisle');

INSERT INTO item (category_id, name) VALUES
(1, 'Breakfast sausage'),
(1, 'Tofu'),
(1, 'Pizza'),
(2, 'Cashews'),
(2, 'Garlic powder'),
(2, 'Almonds'),
(1, 'Jolly Llama'),
(2, 'Dried beets');

INSERT INTO item_list (item_id, quantity, done) VALUES
(1, '1 package', FALSE),
(2, '', FALSE),
(3, '2', FALSE),
(4, '1 cup', TRUE),
(5, '1.5oz', FALSE),
(6, '0.5lb', FALSE);
-- +goose StatementEnd

-- +goose Down
