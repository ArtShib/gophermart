-- +goose Up
-- +goose StatementBegin
create table if not exists orders
(
    id          bigserial PRIMARY KEY,
    number      bigint not null UNIQUE,
    status      text default 'NEW',
    accrual     float8 default null,
    uploaded_at bigint not null,
    user_id     bigint not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table orders;
-- +goose StatementEnd
