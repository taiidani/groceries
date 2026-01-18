-- +goose Up
-- +goose StatementBegin
DELETE FROM "recipe_item";
ALTER SEQUENCE "recipe_item_id_seq" RESTART WITH 1;
DELETE FROM "recipe";
ALTER SEQUENCE "recipe_id_seq" RESTART WITH 1;
DELETE FROM item_bag;
ALTER SEQUENCE item_bag_id_seq RESTART WITH 1;
DELETE FROM item_list;
ALTER SEQUENCE item_list_id_seq RESTART WITH 1;
DELETE FROM item;
ALTER SEQUENCE item_id_seq RESTART WITH 1;
DELETE FROM category;
ALTER SEQUENCE category_id_seq RESTART WITH 1;
DELETE FROM "store";
ALTER SEQUENCE "store_id_seq" RESTART WITH 1;
DELETE FROM "user_group";
ALTER SEQUENCE "user_group_id_seq" RESTART WITH 1;
DELETE FROM "group";
ALTER SEQUENCE "group_id_seq" RESTART WITH 1;
DELETE FROM "user";
ALTER SEQUENCE user_id_seq RESTART WITH 1;
-- +goose StatementEnd

-- +goose Down
