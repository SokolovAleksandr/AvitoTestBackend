-- +goose Up
-- +goose StatementBegin
CREATE TABLE users(id BIGINT PRIMARY KEY);
CREATE TABLE balances(
    userId BIGINT PRIMARY KEY REFERENCES users (id),
    value BIGINT
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE balances;
DROP TABLE users;
-- +goose StatementEnd