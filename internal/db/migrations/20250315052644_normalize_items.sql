-- +goose Up
-- +goose StatementBegin
CREATE TABLE item_bag (
    id SERIAL PRIMARY KEY,
    item_id INTEGER NOT NULL UNIQUE REFERENCES item,
    quantity VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE item_list (
    id SERIAL PRIMARY KEY,
    item_id INTEGER NOT NULL UNIQUE REFERENCES item,
    quantity VARCHAR(255) NOT NULL DEFAULT '',
    done BOOLEAN NOT NULL DEFAULT FALSE
);

INSERT INTO item_bag (item_id, quantity) SELECT id, quantity FROM item WHERE in_bag IS TRUE;
INSERT INTO item_list (item_id, quantity, done) SELECT id, quantity, done FROM item WHERE in_bag IS FALSE;

ALTER TABLE item DROP COLUMN quantity;
ALTER TABLE item DROP COLUMN in_bag;
ALTER TABLE item DROP COLUMN done;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE item ADD COLUMN quantity VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE item ADD COLUMN in_bag BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE item ADD COLUMN done BOOLEAN NOT NULL DEFAULT FALSE;

DROP TABLE item_list;
DROP TABLE item_bag;
DROP TABLE item_recipe;
DROP TABLE recipe;
-- +goose StatementEnd
