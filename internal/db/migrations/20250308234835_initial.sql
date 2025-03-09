-- +goose Up
CREATE TABLE category (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE item (
    id SERIAL PRIMARY KEY,
    category_id INTEGER NOT NULL REFERENCES category,
    name VARCHAR(255) NOT NULL UNIQUE,
    quantity VARCHAR(255),
    in_bag BOOLEAN NOT NULL DEFAULT FALSE,
    done BOOLEAN NOT NULL DEFAULT FALSE
);

-- +goose Down
DROP TABLE item;
DROP TABLE category;
