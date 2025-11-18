-- +goose Up
-- +goose StatementBegin
create or replace view balance as (
with
    withdrawn as (select
				    w.user_id,
					sum(COALESCE(w.sum, 0)) as withdrawn
				  from withdrawal_accruals w
				  group by w.user_id),
	accrual as (select
				    o.user_id,
					sum(COALESCE(o.accrual, 0)) as accrual
				from orders o
				group by o.user_id)

select
    a.user_id,
	a.accrual - COALESCE(w.withdrawn, 0) as current,
	COALESCE(w.withdrawn, 0) as withdrawn
from accrual a
left join withdrawn w on w.user_id = a.user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop view balance;
-- +goose StatementEnd
