-- +goose Up
-- +goose StatementBegin
INSERT INTO store (id, name) VALUES (0, 'Uncategorized');
UPDATE category SET store_id = 0 WHERE store_id IS NULL;
ALTER TABLE "category" ALTER COLUMN store_id SET NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "category" ALTER COLUMN store_id DROP NOT NULL;
UPDATE category SET store_id = NULL WHERE store_id = 0;
DELETE FROM store WHERE id = 0;
-- +goose StatementEnd
