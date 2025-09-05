package yatracker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/corray333/tg-task-parser/internal/config"
	"github.com/corray333/tg-task-parser/internal/entities/task"
)

type YaTrackerRepository struct {
	client *http.Client
}

type CreateTaskRequest struct {
	Summary     string  `json:"summary"`
	Queue       int     `json:"queue"`
	Description string  `json:"description"`
	Project     Project `json:"project"`
}

type Project struct {
	Primary int `json:"primary"`
}

type CreateTaskResponse struct {
	Self    string `json:"self"`
	Key     string `json:"key"`
	Summary string `json:"summary"`
}

// Issue описывает минимальный набор полей задачи, который нам нужен при поиске
type Issue struct {
	Self    string `json:"self"`
	Key     string `json:"key"`
	Summary string `json:"summary"`
}

// SearchIssuesRequest — тело запроса для поиска задач
type SearchIssuesRequest struct {
	Filter  map[string]interface{} `json:"filter,omitempty"`
	PerPage int                    `json:"perPage,omitempty"`
	Order   string                 `json:"order,omitempty"`
}

func NewYaTrackerRepository() *YaTrackerRepository {
	return &YaTrackerRepository{
		client: &http.Client{},
	}
}

func (r *YaTrackerRepository) CreateTask(ctx context.Context, task *task.Task) (*Issue, error) {
	cfg := config.GetYaTrackerConfig()
	token := config.GetYaTrackerToken()

	// Создаем запрос
	reqBody := CreateTaskRequest{
		Summary: task.Title,
		Queue:   cfg.QueueID,
		Project: Project{
			Primary: cfg.ProjectID,
		},
		Description: task.PlainBody,
	}

	// Сериализуем в JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Создаем HTTP запрос
	req, err := http.NewRequestWithContext(ctx, "POST", cfg.APIURL+"/issues", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Authorization", "OAuth "+token)
	req.Header.Set("X-Cloud-Org-Id", cfg.OrgID)
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Парсим ответ в объект Issue
	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &issue, nil
}

// SearchTasksByName выполняет поиск задач по имени (summary),
// используя текст из переданной задачи (task.Text)
func (r *YaTrackerRepository) SearchTasksByName(ctx context.Context, t *task.Task) ([]Issue, error) {
	slog.Info("Search task in yandex tracker", "task", t)
	cfg := config.GetYaTrackerConfig()
	token := config.GetYaTrackerToken()

	// Формируем тело запроса: фильтрация по полю summary
	searchReq := SearchIssuesRequest{
		Filter: map[string]interface{}{
			"summary": t.Title,
		},
	}

	jsonData, err := json.Marshal(searchReq)
	if err != nil {
		slog.Error("Failed to marshal search request in yandex tracker", "error", err)
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cfg.APIURL+"/issues/_search?perPage=50", bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error("Failed to create search request in yandex tracker", "error", err)
		return nil, fmt.Errorf("failed to create search request: %w", err)
	}

	req.Header.Set("Authorization", "OAuth "+token)
	req.Header.Set("X-Cloud-Org-Id", cfg.OrgID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		slog.Error("Failed to execute search request in yandex tracker", "error", err)
		return nil, fmt.Errorf("failed to execute search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("Unexpected status code while searching issue in yandex tracker", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var issues []Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		slog.Error("Failed to decode search issues in yandex tracker response", "error", err)
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	slog.Info("Found task in yandex tracker", "issues", issues)

	return issues, nil
}
