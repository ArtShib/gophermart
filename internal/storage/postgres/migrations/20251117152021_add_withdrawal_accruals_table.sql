-- +goose Up
-- +goose StatementBegin
create table if not exists withdrawal_accruals
(
    id              bigserial PRIMARY KEY,
    order_id        bigint not null,
    user_id         bigint not null ,
    sum             float8 not null,
    processed_at    bigint not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table withdrawal_accruals;
-- +goose StatementEnd
