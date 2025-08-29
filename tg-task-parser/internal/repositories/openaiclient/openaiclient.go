package openaiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/corray333/tg-task-parser/internal/entities/task"
	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

type OpenAIRepository struct {
	client *openai.Client
	prompt string
	model  string
}

// NewOpenAIRepository creates a new OpenAI repository with configuration from environment variables and viper
func NewOpenAIRepository() (*OpenAIRepository, error) {
	// Get API key from environment variable (sensitive data)
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		slog.Error("OPENAI_API_KEY environment variable is not set")
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	// Get non-sensitive configuration from viper
	model := viper.GetString("openai.model")
	if model == "" {
		slog.Error("openai.model configuration is not set")
		return nil, fmt.Errorf("openai.model configuration is required")
	}

	prompt := viper.GetString("openai.prompt")
	if prompt == "" {
		slog.Error("openai.prompt configuration is not set")
		return nil, fmt.Errorf("openai.prompt configuration is required")
	}

	client := openai.NewClient(apiKey)

	slog.Info("OpenAI repository initialized", "model", model, "prompt", prompt)

	return &OpenAIRepository{
		client: client,
		prompt: prompt,
		model:  model,
	}, nil
}

// ParseMessage sends a message to OpenAI API and parses the response into a Task struct
func (r *OpenAIRepository) ParseMessage(ctx context.Context, message string) (*task.Task, error) {
	// Validate input message
	if strings.TrimSpace(message) == "" {
		slog.Error("Empty message provided for parsing")
		return nil, fmt.Errorf("message cannot be empty")
	}

	slog.Info("Starting message parsing", "message_length", len(message), "model", r.model)

	// Prepare the prompt with message
	content := fmt.Sprintf("%s\n\nMessage: %s", r.prompt, message)

	req := openai.ChatCompletionRequest{
		Model: r.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: content,
			},
		},
		// MaxTokens: 1000,
	}

	slog.Info("Sending request to OpenAI API")

	resp, err := r.client.CreateChatCompletion(ctx, req)
	if err != nil {
		slog.Error("Failed to create chat completion", "error", err)
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		slog.Error("OpenAI response contains no choices")
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	content = resp.Choices[0].Message.Content
	slog.Info("Received response from OpenAI", "content_length", len(content), "usage_tokens", resp.Usage.TotalTokens)

	// Parse the content as Task JSON
	var taskResult task.Task
	if err := json.Unmarshal([]byte(content), &taskResult); err != nil {
		slog.Error("Failed to unmarshal task from OpenAI response", "error", err, "content", content)
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	slog.Info("Successfully parsed task",
		"task", taskResult,
		"body", string(taskResult.Body),
	)

	return &taskResult, nil
}
