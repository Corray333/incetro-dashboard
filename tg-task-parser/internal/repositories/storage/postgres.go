package storage

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository() *Repository {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB_NAME"))
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	if err := goose.Up(db.DB, "../migrations"); err != nil {
		panic(err)
	}

	return &Repository{
		db: db,
	}
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
	fmt.Println(chatID)
	if err := r.db.Get(&projectID, "SELECT project_id FROM chats WHERE chat_id = $1", chatID); err != nil {
		slog.Error("Error while getting project by chat ID", "error", err)
		return uuid.Nil, err
	}
	return projectID, nil
}
