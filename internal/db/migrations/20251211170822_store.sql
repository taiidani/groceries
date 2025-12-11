-- +goose Up
-- +goose StatementBegin
CREATE TABLE "store" (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

ALTER TABLE "category" ADD COLUMN "store_id" INTEGER REFERENCES "store" (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "category" DROP COLUMN "store_id";
DROP TABLE IF EXISTS "store";
-- +goose StatementEnd
