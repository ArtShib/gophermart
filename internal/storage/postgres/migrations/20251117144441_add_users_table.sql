-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users
(
    id        bigserial PRIMARY KEY,
    login     text NOT NULL UNIQUE,
    pass_hash text NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table users;
-- +goose StatementEnd
