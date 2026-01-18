-- +goose Up
CREATE TABLE recipe (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE recipe_item (
    id SERIAL PRIMARY KEY,
    recipe_id INTEGER NOT NULL REFERENCES recipe(id) ON DELETE CASCADE,
    item_id INTEGER NOT NULL REFERENCES item(id) ON DELETE RESTRICT,
    quantity VARCHAR(255),
    UNIQUE(recipe_id, item_id)
);

CREATE INDEX idx_recipe_item_recipe_id ON recipe_item(recipe_id);
CREATE INDEX idx_recipe_item_item_id ON recipe_item(item_id);

-- +goose Down
-- +goose StatementBegin
DROP TABLE recipe_item;
DROP TABLE recipe;
-- +goose StatementEnd
