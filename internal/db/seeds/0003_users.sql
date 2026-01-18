-- +goose Up
-- +goose StatementBegin

INSERT INTO "user" (id, name, admin) VALUES
(1, 'admin', TRUE),
(2, 'user', FALSE);

INSERT INTO "user_group" (user_id, group_id) VALUES (1, 1);
INSERT INTO "user_group" (user_id, group_id) VALUES (2, 2);

-- +goose StatementEnd

-- +goose Down
