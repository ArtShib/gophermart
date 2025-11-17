-- +goose Up
-- +goose StatementBegin
create or replace trigger check_balance_user
before insert or update on withdrawal_accruals
for each row execute function check_balance_user();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop trigger check_balance_user on withdrawal_accruals;
-- +goose StatementEnd
