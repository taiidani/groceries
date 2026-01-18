-- +goose Up
-- +goose StatementBegin
INSERT INTO recipe (name, description) VALUES
('Spaghetti Carbonara', 'Classic Italian pasta dish with eggs, cheese, and pancetta');

-- Recipe items (using items that exist: Free will=1, Love & Peace=2, Tofu=5)
INSERT INTO recipe_item (recipe_id, item_id, quantity) VALUES
(1, 1, '1 cup'),
(1, 2, '2 tbsp'),
(1, 5, '200g');
-- +goose StatementEnd

-- +goose Down
