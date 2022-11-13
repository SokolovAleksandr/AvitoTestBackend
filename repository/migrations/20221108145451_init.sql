-- +goose Up
-- +goose StatementBegin
CREATE TABLE users(id BIGINT PRIMARY KEY);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd