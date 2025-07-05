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
('Produce', 'Only the freshest'),
('Bulk Foods', 'Mostly nuts'),
('Exotic Pets', 'Not a frequented aisle'),
('Household Items', '');

INSERT INTO item (category_id, name) VALUES
(1, 'Breakfast sausage'),
(1, 'Tofu'),
(1, 'Pizza'),
(2, 'Cashews'),
(2, 'Garlic powder'),
(2, 'Almonds'),
(1, 'Jolly Llama'),
(2, 'Dried beets'),
(4, 'Shower curtain'),
(4, 'Plastic storage bins'),
(4, 'Closet organizers '),
(4, 'Garbage bags'),
(4, 'Veggie scrubber'),
(4, 'Dishwasher tabs'),
(4, 'Liquid Castile soap'),
(4, 'Plant sprayer'),
(1, 'Lemons'),
(1, 'Mangoes'),
(1, 'Tangerines'),
(1, 'Apples'),
(1, 'Bell pepper'),
(1, 'Scallions'),
(1, 'Cucumber'),
(1, 'Parsley'),
(1, 'Sweet potatoes'),
(1, 'Little potatoes'),
(1, 'Onion'),
(1, 'Broccoli'),
(1, 'Shallots'),
(1, 'Ginger'),
(1, 'Garlic'),
(1, 'Bananas'),
(1, 'Crimini Mushrooms'),
(1, 'Carrots'),
(1, 'Salad Greens'),
(1, 'Strawberries'),
(1, 'Serrano chiles'),
(1, 'Lemongrass stalks'),
(1, 'Galangal'),
(1, 'Lime'),
(1, 'Small eggplants'),
(1, 'Zucchini'),
(1, 'Snow peas'),
(1, 'Cherry tomatoes'),
(1, 'Spinach'),
(1, 'Fruit'),
(1, 'Blueberries'),
(1, 'Lacinato kale'),
(1, 'Arugula'),
(1, 'Avocado'),
(1, 'Coleslaw mix'),
(1, 'Peaches'),
(1, 'Napa cabbage'),
(1, 'Basil'),
(1, 'Chives'),
(1, 'Dill'),
(1, 'Green beans'),
(1, 'Baby spinach'),
(1, 'Cilantro'),
(1, 'Mushrooms'),
(1, 'Corn'),
(1, 'Russet potatoes'),
(1, 'Romaine lettuce');

INSERT INTO item_list (item_id, quantity, done) VALUES
(1, '1 package', FALSE),
(2, '', FALSE),
(3, '2', FALSE),
(4, '1 cup', TRUE),
(5, '1.5oz', FALSE),
(6, '0.5lb', FALSE);
-- +goose StatementEnd

-- +goose Down
