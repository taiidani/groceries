-- +goose Up
-- +goose StatementBegin

INSERT INTO "group" (name) VALUES ('Smiths');
INSERT INTO "group" (name) VALUES ('Jones');
-- +goose StatementEnd

-- +goose Down
