package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/Corray333/employee_dashboard/internal/errs"
	"github.com/jmoiron/sqlx"
)

type TxKey struct{}

type PostgresClient struct {
	db *sqlx.DB
	// courses map[string]string
}

func (s *PostgresClient) DB() *sqlx.DB {
	return s.db
}

type Transactioner interface {
	Begin(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	GetTx(ctx context.Context) (tx *sqlx.Tx, isNew bool, err error)
}

func New() *PostgresClient {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB_NAME"))
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}

	s := &PostgresClient{
		db: db,
	}

	return s
}

func (r *PostgresClient) Begin(ctx context.Context) (context.Context, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		slog.Error("Failed to begin transaction: " + err.Error())
		return nil, err
	}

	return context.WithValue(ctx, TxKey{}, tx), nil
}

func (r *PostgresClient) Commit(ctx context.Context) error {
	tx, ok := ctx.Value(TxKey{}).(*sqlx.Tx)
	if !ok {
		slog.Error("Invalid transaction in context")
		return nil
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Failed to commit transaction: " + err.Error())
		return err
	}

	return nil
}

func (r *PostgresClient) Rollback(ctx context.Context) error {
	tx, ok := ctx.Value(TxKey{}).(*sqlx.Tx)
	if !ok {
		slog.Error("Invalid transaction in context")
		return nil
	}

	if err := tx.Rollback(); err != nil {
		slog.Error("Failed to rollback transaction: " + err.Error())
		return err
	}

	return nil
}

func (r *PostgresClient) GetTx(ctx context.Context) (tx *sqlx.Tx, isNew bool, err error) {
	txRaw := ctx.Value(TxKey{})
	if txRaw != nil {
		var ok bool
		tx, ok = txRaw.(*sqlx.Tx)
		if !ok {
			slog.Error("invalid transaction type")
			return nil, false, errs.ErrInvalidTxTypeInCtx
		}
	}
	if tx == nil {
		tx, err = r.db.BeginTxx(ctx, nil)
		if err != nil {
			slog.Error("Failed to begin transaction", "error", err)
			return nil, false, err
		}

		return tx, true, nil
	}

	return tx, false, nil
}
