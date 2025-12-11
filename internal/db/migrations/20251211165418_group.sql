-- +goose Up
-- +goose StatementBegin
CREATE TABLE "group" (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE "user_group" (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES "user" (id),
    group_id INTEGER NOT NULL REFERENCES "group" (id),
    UNIQUE (user_id, group_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "user_group";
DROP TABLE IF EXISTS "group";
-- +goose StatementEnd
