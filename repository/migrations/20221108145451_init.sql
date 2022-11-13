-- +goose Up
-- +goose StatementBegin
CREATE TABLE users(id BIGINT PRIMARY KEY);
CREATE TABLE balances(
    userId BIGINT PRIMARY KEY REFERENCES users (id),
    value BIGINT
);
CREATE TABLE reserves(
    id BIGSERIAL PRIMARY KEY,
    userId BIGINT REFERENCES users (id),
    size BIGINT
);
CREATE INDEX user_reserves ON reserves USING hash (userId);
CREATE TABLE expanses(
    id BIGSERIAL PRIMARY KEY,
    fromId BIGINT REFERENCES users (id),
    toId BIGINT REFERENCES users (id),
    ts TIMESTAMP,
    serviceId BIGINT,
    orderId BIGINT,
    cost BIGINT,
    status SMALLINT
);
CREATE INDEX user_expanses ON expanses USING hash (fromId);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX user_expanses;
DROP TABLE expanses;
DROP INDEX user_reserves;
DROP TABLE reserves;
DROP TABLE balances;
DROP TABLE users;
-- +goose StatementEnd