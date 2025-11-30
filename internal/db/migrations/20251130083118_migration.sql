-- +goose Up
-- +goose StatementBegin
INSERT INTO "user" (id, name, admin) VALUES (1, 'taiidani', TRUE);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM "user" WHERE id = 1;
-- +goose StatementEnd
