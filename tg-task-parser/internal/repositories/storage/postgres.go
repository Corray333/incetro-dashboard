package storage

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/corray333/tg-task-parser/internal/entities/message"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ErrInvalidTxType = fmt.Errorf("invalid transaction type")

type Repository struct {
	db *sqlx.DB
}

type TxKey struct{}

func NewRepository() *Repository {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB_NAME"))
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	// if err := goose.SetDialect("postgres"); err != nil {
	// 	panic(err)
	// }

	// if err := goose.Up(db.DB, "../migrations"); err != nil {
	// 	panic(err)
	// }

	return &Repository{
		db: db,
	}
}

func (r *Repository) Begin(ctx context.Context) (context.Context, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, TxKey{}, tx), nil
}

func (r *Repository) Commit(ctx context.Context) error {
	tx, ok := ctx.Value(TxKey{}).(*sqlx.Tx)
	if !ok {
		return nil
	}

	return tx.Commit()
}

func (r *Repository) Rollback(ctx context.Context) error {
	tx, ok := ctx.Value(TxKey{}).(*sqlx.Tx)
	if !ok {
		return nil
	}

	return tx.Rollback()
}

func (r *Repository) getTx(ctx context.Context) (tx *sqlx.Tx, isNew bool, err error) {
	txRaw := ctx.Value(TxKey{})
	if txRaw != nil {
		var ok bool
		tx, ok = txRaw.(*sqlx.Tx)
		if !ok {
			slog.Error("invalid transaction type")
			return nil, false, ErrInvalidTxType
		}
	}
	if tx == nil {
		tx, err = r.db.BeginTxx(ctx, nil)
		if err != nil {
			slog.Error("failed to begin transaction", "error", err)
			return nil, false, err
		}

		return tx, true, nil
	}

	return tx, false, nil
}

func (r *Repository) LinkChatToProject(ctx context.Context, chatID int64, projectID uuid.UUID) error {
	if _, err := r.db.Exec("INSERT INTO chats (chat_id, project_id) VALUES ($1, $2)", chatID, projectID); err != nil {
		slog.Error("Error while linking chat to project", "error", err)
		return err
	}
	return nil
}

func (r *Repository) GetProjectByChatID(ctx context.Context, chatID int64) (uuid.UUID, error) {
	var projectID uuid.UUID
	if err := r.db.Get(&projectID, "SELECT project_id FROM chats WHERE chat_id = $1", chatID); err != nil {
		slog.Error("Error while getting project by chat ID", "error", err)
		return uuid.Nil, err
	}
	return projectID, nil
}

func (r *Repository) SaveMessage(ctx context.Context, message message.Message) error {
	if _, err := r.db.Exec("INSERT INTO tg_messages (chat_id, message_id, text) VALUES ($1, $2, $3)", message.ChatID, message.MessageID, message.Text); err != nil {
		slog.Error("Error while saving message", "error", err)
		return err
	}
	return nil
}
