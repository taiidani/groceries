-- +goose Up
CREATE TABLE category (
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE item (
    category_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL UNIQUE,
    quantity VARCHAR(255),
    in_bag BOOLEAN NOT NULL DEFAULT FALSE,
    done BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY(category_id) REFERENCES category(id)
);

-- +goose Down
DROP TABLE item;
DROP TABLE category;
