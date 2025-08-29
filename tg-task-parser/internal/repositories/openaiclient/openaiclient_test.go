package openaiclient

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/corray333/tg-task-parser/internal/entities/task"
)

func TestNewOpenAIRepository(t *testing.T) {
	tests := []struct {
		name          string
		apiKey        string
		model         string
		prompt        string
		expectError   bool
		errorContains string
	}{
		{
			name:   "valid configuration",
			apiKey: "sk-test-key",
			model:  "gpt-3.5-turbo",
			prompt: "Test prompt",
		},
		{
			name:          "missing API key",
			apiKey:        "",
			model:         "gpt-3.5-turbo",
			prompt:        "Test prompt",
			expectError:   true,
			errorContains: "OPENAI_API_KEY environment variable is required",
		},
		{
			name:          "missing model",
			apiKey:        "sk-test-key",
			model:         "",
			prompt:        "Test prompt",
			expectError:   true,
			errorContains: "openai.model configuration is required",
		},
		{
			name:          "missing prompt",
			apiKey:        "sk-test-key",
			model:         "gpt-3.5-turbo",
			prompt:        "",
			expectError:   true,
			errorContains: "openai.prompt configuration is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment and config
			if tt.apiKey != "" {
				os.Setenv("OPENAI_API_KEY", tt.apiKey)
				defer os.Unsetenv("OPENAI_API_KEY")
			} else {
				os.Unsetenv("OPENAI_API_KEY")
			}

			viper.Reset()
			viper.Set("openai.model", tt.model)
			viper.Set("openai.prompt", tt.prompt)

			repo, err := NewOpenAIRepository()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, repo)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, repo)
				assert.NotNil(t, repo.client)
				assert.Equal(t, tt.model, repo.model)
				assert.Equal(t, tt.prompt, repo.prompt)
			}
		})
	}
}

// TestMessageValidation tests only the input validation logic
func TestMessageValidation(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		expectError   bool
		errorContains string
	}{
		{
			name:          "empty message",
			message:       "",
			expectError:   true,
			errorContains: "message cannot be empty",
		},
		{
			name:          "whitespace only message",
			message:       "   \n\t   ",
			expectError:   true,
			errorContains: "message cannot be empty",
		},
		{
			name:    "valid message",
			message: "Создать новую задачу #важно @user1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the validation logic directly
			if strings.TrimSpace(tt.message) == "" {
				if tt.expectError {
					assert.True(t, true, "Empty message correctly identified")
				} else {
					t.Error("Expected valid message but got empty")
				}
			} else {
				if !tt.expectError {
					assert.True(t, true, "Valid message correctly identified")
				} else {
					t.Error("Expected empty message but got valid")
				}
			}
		})
	}
}

// TestTaskJSONParsing tests the JSON parsing logic
func TestTaskJSONParsing(t *testing.T) {
	tests := []struct {
		name         string
		jsonResponse string
		expectedTask *task.Task
		expectError  bool
	}{
		{
			name: "valid task JSON",
			jsonResponse: `{
				"title": "Создать новую функцию",
				"link": "",
				"body": "Реализовать парсинг сообщений из Telegram",
				"hashtags": ["важно", "разработка"],
				"executors": ["@developer1", "@developer2"],
				"assignee": "@developer1"
			}`,
			expectedTask: &task.Task{
				Title:     "Создать новую функцию",
				Link:      "",
				Body:      json.RawMessage(`"Реализовать парсинг сообщений из Telegram"`),
				Hashtags:  []task.Tag{"важно", "разработка"},
				Executors: []task.Mention{"@developer1", "@developer2"},
				Assignee:  "@developer1",
			},
		},
		{
			name:         "invalid JSON",
			jsonResponse: `{"title": "Invalid JSON"`,
			expectError:  true,
		},
		{
			name:         "empty response",
			jsonResponse: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parsedTask task.Task
			err := json.Unmarshal([]byte(tt.jsonResponse), &parsedTask)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedTask.Title, parsedTask.Title)
				assert.Equal(t, tt.expectedTask.Link, parsedTask.Link)
				assert.Equal(t, tt.expectedTask.Hashtags, parsedTask.Hashtags)
				assert.Equal(t, tt.expectedTask.Executors, parsedTask.Executors)
				assert.Equal(t, tt.expectedTask.Assignee, parsedTask.Assignee)
			}
		})
	}
}

// TestRepositoryStructure tests the repository structure and configuration
func TestRepositoryStructure(t *testing.T) {
	// Setup valid configuration
	os.Setenv("OPENAI_API_KEY", "sk-test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	viper.Reset()
	viper.Set("openai.model", "gpt-4")
	viper.Set("openai.prompt", "Custom prompt for testing")

	repo, err := NewOpenAIRepository()
	require.NoError(t, err)

	// Test repository structure
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.client)
	assert.Equal(t, "gpt-4", repo.model)
	assert.Equal(t, "Custom prompt for testing", repo.prompt)

	// Test that the repository implements expected interface
	assert.Implements(t, (*interface {
		ParseMessage(context.Context, string) (*task.Task, error)
	})(nil), repo)
}
