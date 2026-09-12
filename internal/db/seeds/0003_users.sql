-- +goose Up
-- +goose StatementBegin

INSERT INTO "user" (id, name, admin) VALUES
(1, 'admin', TRUE),
(2, 'user', FALSE);

INSERT INTO "user_group" (user_id, group_id) VALUES (1, 1);
INSERT INTO "user_group" (user_id, group_id) VALUES (2, 2);

-- Explicit ids above don't advance the sequence (unlike store/category,
-- which reserve id=0 for their special row and let the sequence handle the
-- rest); bump it so subsequently created users (e.g. via OIDC
-- auto-provisioning) don't collide with these seeded rows.
SELECT setval('user_id_seq', (SELECT MAX(id) FROM "user"));

-- +goose StatementEnd

-- +goose Down
