package yatracker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/corray333/tg-task-parser/internal/config"
	"github.com/corray333/tg-task-parser/internal/entities/task"
)

type YaTrackerRepository struct {
	client *http.Client
}

type CreateTaskRequest struct {
	Summary string  `json:"summary"`
	Queue   int     `json:"queue"`
	Project Project `json:"project"`
}

type Project struct {
	Primary int `json:"primary"`
}

type CreateTaskResponse struct {
	Self    string `json:"self"`
	Key     string `json:"key"`
	Summary string `json:"summary"`
}

func NewYaTrackerRepository() *YaTrackerRepository {
	return &YaTrackerRepository{
		client: &http.Client{},
	}
}

func (r *YaTrackerRepository) CreateTas(ctx context.Context, task *task.Task) (string, string, error) {
	cfg := config.GetYaTrackerConfig()
	token := config.GetYaTrackerToken()

	// Создаем запрос
	reqBody := CreateTaskRequest{
		Summary: task.Text,
		Queue:   cfg.QueueID,
		Project: Project{
			Primary: cfg.ProjectID,
		},
	}

	// Сериализуем в JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Создаем HTTP запрос
	req, err := http.NewRequestWithContext(ctx, "POST", cfg.APIURL+"/issues", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Authorization", "OAuth "+token)
	req.Header.Set("X-Cloud-Org-Id", cfg.OrgID)
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	resp, err := r.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusCreated {
		return "", "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Парсим ответ
	var response CreateTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Формируем название задачи в формате "key: summary"
	taskTitle := fmt.Sprintf("%s: %s", response.Key, response.Summary)

	return response.Self, taskTitle, nil
}
