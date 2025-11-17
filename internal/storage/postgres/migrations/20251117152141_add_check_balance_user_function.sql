-- +goose Up
-- +goose StatementBegin
create or replace function check_balance_user()
returns trigger as $$
declare
    current_balance float8;
	order_id bigint;
begin
SELECT "current" INTO current_balance
FROM balance
WHERE user_id = NEW.user_id
    FOR UPDATE;
if COALESCE(current_balance, 0) <= new.sum then
    raise exception 'there are not enough bonuses to deduct';
ELSE
	INSERT INTO orders (number, user_id, uploaded_at)
    VALUES (NEW.order_id, NEW.user_id, NEW.processed_at) RETURNING id INTO order_id;
	NEW.order_id = order_id;
end if;
return new;
end;
$$ language plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop function check_balance_user();
-- +goose StatementEnd
