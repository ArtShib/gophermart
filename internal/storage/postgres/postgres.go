package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ArtShib/gophermart.git/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type StorePostgres struct {
	db *sql.DB
}

//go:embed migrations/*.sql
var migrations embed.FS

func NewPostgresStore(ctx context.Context, connectionString string) (*StorePostgres, error) {
	const op = "storage.postgres.NewPostgresStore"

	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	nCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	goose.SetBaseFS(migrations)
	if err := goose.UpContext(nCtx, db, "migrations"); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &StorePostgres{db: db}, nil
}

func (pg *StorePostgres) Close() error {
	return pg.db.Close()
}

func (pg *StorePostgres) SaveUser(ctx context.Context, login string, passHash []byte) (*models.User, error) {
	const op = "storage.postgres.SaveUser"
	var user models.User
	stmt, err := pg.db.Prepare("INSERT INTO users (login, pass_hash) VALUES ($1, $2) RETURNING id, login, pass_hash")

	if err != nil {
		return &models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	err = stmt.QueryRowContext(ctx, login, passHash).Scan(&user.ID, &user.Login, &user.PassHash)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return &models.User{}, fmt.Errorf("%s: %w", op, models.ErrUserExists)
			}
		}
		return &models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (pg *StorePostgres) User(ctx context.Context, login string) (*models.User, error) {
	const op = "storage.postgres.User"

	stmt, err := pg.db.Prepare("SELECT id, login, pass_hash FROM users WHERE login = $1")
	if err != nil {
		return &models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, login)

	var user models.User
	err = row.Scan(&user.ID, &user.Login, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.User{}, fmt.Errorf("%s: %w", op, models.ErrUserNotFound)
		}
		return &models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (pg *StorePostgres) AddOrder(ctx context.Context, numOrder int64, uploaded int64, userID int64) error {
	const op = "storage.postgres.AddOrder"
	stmt, err := pg.db.Prepare("INSERT INTO orders (number, uploaded_at, user_id) VALUES ($1, $2, $3)")

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, numOrder, uploaded, userID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				var userExists int64
				stmtExists, err := pg.db.Prepare("SELECT user_id FROM orders WHERE number = $1")
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}
				if err := stmtExists.QueryRowContext(ctx, numOrder).Scan(&userExists); err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}
				if userExists == userID {
					return fmt.Errorf("%s: %w", op, models.ErrOrderExists)
				}
				return fmt.Errorf("%s: %w", op, models.ErrOrderExistsOtherUser)
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (pg *StorePostgres) GetOrder(ctx context.Context, userID int64) (models.OrderArray, error) {
	const op = "storage.postgres.GetOrder"
	stmt, err := pg.db.Prepare("select number, status, accrual, uploaded_at from orders where user_id = $1 order by uploaded_at desc")

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error(op, "Error", err)
		}
	}()

	orderArray := models.OrderArray{}
	for rows.Next() {
		var order models.Order
		var status sql.NullString
		var accrual sql.NullFloat64

		if err := rows.Scan(&order.Number, &status, &accrual, &order.UploadedAt); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		order.Status = status.String
		order.Accrual = accrual.Float64
		orderArray = append(orderArray, order)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(orderArray) == 0 {
		return nil, fmt.Errorf("%s: %w", op, models.ErrOrderExistsOtherUser)
	}
	return orderArray, nil
}

func (pg *StorePostgres) GetBalance(ctx context.Context, userID int64) (*models.Balance, error) {
	const op = "storage.postgres.GetBalance"
	stmt, err := pg.db.Prepare("select current, withdrawn from balance where user_id = $1")

	if err != nil {
		return &models.Balance{}, fmt.Errorf("%s: %w", op, err)
	}

	var current, withdrawn sql.NullFloat64
	err = stmt.QueryRowContext(ctx, userID).Scan(&current, &withdrawn)

	if err != nil {
		return &models.Balance{}, fmt.Errorf("%s: %w", op, err)
	}

	balance := &models.Balance{}
	balance.Current = current.Float64
	balance.Withdrawn = withdrawn.Float64

	return balance, nil
}

func (pg *StorePostgres) GetWithdrawals(ctx context.Context, userID int64) (models.WithdrawalsArray, error) {
	const op = "storage.postgres.GetWithdrawals"
	stmt, err := pg.db.Prepare(`select
											o.number,
											w.sum,
											w.processed_at
										from withdrawal_accruals w
										left join orders o on o.id = w.order_id
										where w.user_id = $1;`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(ctx, userID)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error(op, "Error", err)
		}
	}()

	withdrawalsArray := models.WithdrawalsArray{}
	for rows.Next() {
		var number sql.NullInt64
		var sum sql.NullFloat64
		var processed sql.NullInt64

		if err := rows.Scan(&number, &sum, &processed); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		withdrawals := models.Withdrawals{
			OrderNum:    number.Int64,
			Sum:         sum.Float64,
			ProcessedAt: processed.Int64,
		}
		withdrawalsArray = append(withdrawalsArray, withdrawals)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(withdrawalsArray) == 0 {
		return nil, fmt.Errorf("%s: %w", op, models.ErrWithdrawalsEmpty)
	}
	return withdrawalsArray, nil
}

func (pg *StorePostgres) AddWithdraw(ctx context.Context, numOrder int64, userID int64, sum float64, processed int64) error {
	const op = "storage.postgres.AddWithdrawal"
	stmt, err := pg.db.Prepare(`
									insert into withdrawal_accruals(order_id, user_id, sum, processed_at)
									values ($1, $2, $3, $4);`)
	//values ((select id from orders where number = $1), $2, $3, $4);`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, numOrder, userID, sum, processed)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Message == "there are not enough bonuses to deduct" {
				return fmt.Errorf("%s: %w", op, models.ErrWithdrawBalanceUser)
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (pg *StorePostgres) GetOrdersInWork(ctx context.Context) (models.OrderArray, error) {
	const op = "storage.postgres.GetOrdersInWork"
	stmt, err := pg.db.Prepare("SELECT number, status, accrual, uploaded_at FROM orders WHERE status not in ('INVALID', 'PROCESSED');")

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(ctx)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error(op, "Error", err)
		}
	}()

	orderArray := models.OrderArray{}
	for rows.Next() {
		var order models.Order
		var status sql.NullString
		var accrual sql.NullFloat64

		if err := rows.Scan(&order.Number, &status, &accrual, &order.UploadedAt); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		order.Status = status.String
		order.Accrual = accrual.Float64
		orderArray = append(orderArray, order)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(orderArray) == 0 {
		return nil, fmt.Errorf("%s: %w", op, models.ErrOrdersInWorkIsEmpty)
	}
	return orderArray, nil
}

func (pg *StorePostgres) UpdateOrdersBatch(ctx context.Context, orders models.ResAccrualOrderArray) error {
	const op = "storage.postgres.UpdateOrdersBatch"

	values := make([]string, len(orders))
	args := make([]interface{}, 0, len(orders)*3)
	for i, order := range orders {
		pos1, pos2, pos3 := len(args)+1, len(args)+2, len(args)+3
		values[i] = fmt.Sprintf("($%d::bigint, $%d::text, $%d::float8)", pos1, pos2, pos3)
		args = append(args, order.OrderNum, order.Status, order.Accrual)
	}

	query := fmt.Sprintf(`
        UPDATE orders o
        SET status = v.status,
            accrual = v.accrual
        FROM (VALUES %s) AS v(number, status, accrual)
        WHERE o.number = v.number
    `, strings.Join(values, ", "))

	stmt, err := pg.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
