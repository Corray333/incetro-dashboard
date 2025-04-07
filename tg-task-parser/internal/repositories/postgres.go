package repositories

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/corray333/tg-task-parser/internal/entities/task"
	"github.com/corray333/tg-task-parser/pkg/notion"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

type Repository struct {
	db *sqlx.DB
}

func New() *Repository {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB_NAME"))
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	return &Repository{
		db: db,
	}
}

type PageCreated struct {
	ID string `json:"id"`
}

func (r *Repository) CreateTask(ctx context.Context, task *task.Task) error {
	req := map[string]interface{}{
		"Task": map[string]interface{}{
			"type": "title",
			"title": []map[string]interface{}{
				{
					"type": "text",
					"text": map[string]interface{}{
						"content": task.Text,
					},
				},
			},
		},
		"Исполнитель": map[string]interface{}{
			"type": "people",
			"people": func() []map[string]interface{} {
				executor := []map[string]interface{}{}
				for _, user := range task.Mentions {
					employeeID, err := r.GetEmployeeByTgUsername(ctx, string(user))
					if err != nil {
						slog.Error("Notion error while getting employee ID: " + err.Error())
						return nil
					}
					executor = append(executor, map[string]interface{}{
						"id": employeeID,
					})
				}
				return executor
			}(),
		},
		"Теги": map[string]interface{}{
			"type": "multi_select",
			"multi_select": func() []map[string]interface{} {
				tags := []map[string]interface{}{}
				for _, tag := range task.Hashtags {
					tags = append(tags, map[string]interface{}{
						"name": tag,
					})
				}
				return tags
			}(),
		},
	}

	// Создаем страницу задачи в Notion
	_, err := notion.CreatePage(viper.GetString("notion.databases.tasks"), req, nil, "")
	if err != nil {
		slog.Error("Notion error while creating task: " + err.Error())
		return err
	}

	return nil
}
