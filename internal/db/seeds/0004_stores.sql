-- +goose Up
-- +goose StatementBegin

-- Repeat the row added via the migrations
INSERT INTO store (id, name) VALUES (0, 'Uncategorized');
INSERT INTO store (name) VALUES
('New Seasons'),
('Trader Joe''s');
-- +goose StatementEnd

-- +goose Down
