-- +goose Up
UPDATE item SET quantity = '' WHERE quantity IS NULL;
UPDATE category SET description = '' WHERE description IS NULL;
ALTER TABLE item ALTER COLUMN quantity SET DEFAULT '';
ALTER TABLE category ALTER COLUMN description SET DEFAULT '';
ALTER TABLE category ALTER COLUMN description SET NOT NULL;

-- +goose Down
ALTER TABLE item ALTER COLUMN quantity DROP DEFAULT '';
ALTER TABLE category ALTER COLUMN description DROP DEFAULT '';
ALTER TABLE category ALTER COLUMN description DROP NOT NULL;
