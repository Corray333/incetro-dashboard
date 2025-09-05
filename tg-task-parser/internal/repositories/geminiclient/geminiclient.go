package geminiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/corray333/tg-task-parser/internal/entities/task"
	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
)

type GeminiRepository struct {
	client *genai.Client
}

// NewGeminiRepository creates a new Gemini repository with configuration from environment variables and viper
func NewGeminiRepository() (*GeminiRepository, error) {
	// Get API key from environment variable (sensitive data)
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		slog.Error("GEMINI_API_KEY environment variable is not set")
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}
	slog.Info("Gemini API key loaded successfully", "key", apiKey)

	ctx := context.Background()

	// Create HTTP client with proxy if PROXY_URL is set
	httpClient := &http.Client{}
	proxyURL := os.Getenv("PROXY_URL")
	if proxyURL != "" {
		parsedProxyURL, err := url.Parse(proxyURL)
		if err != nil {
			slog.Error("Failed to parse proxy URL", "proxy_url", proxyURL, "error", err)
			return nil, fmt.Errorf("failed to parse proxy URL: %w", err)
		}
		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(parsedProxyURL),
		}
		slog.Info("Using proxy for Gemini client", "proxy_url", proxyURL)
	}

	// Create Gemini client with custom HTTP client if proxy is configured
	var client *genai.Client
	var err error
	if proxyURL != "" {
		client, err = genai.NewClient(ctx, option.WithAPIKey(apiKey), option.WithHTTPClient(httpClient))
	} else {
		client, err = genai.NewClient(ctx, option.WithAPIKey(apiKey))
	}
	if err != nil {
		slog.Error("Failed to create Gemini client", "error", err)
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	slog.Info("Gemini repository initialized", "model", viper.GetString("gemini.model"), "prompt", viper.GetString("openai.prompt"), "proxy_enabled", proxyURL != "")

	return &GeminiRepository{
		client: client,
	}, nil
}

// ParseMessage sends a message to Gemini API and parses the response into a Task struct
func (r *GeminiRepository) ParseMessage(ctx context.Context, message string) (*task.Task, error) {
	// Validate input message
	if strings.TrimSpace(message) == "" {
		slog.Error("Empty message provided for parsing")
		return nil, fmt.Errorf("message cannot be empty")
	}

	slog.Info("Starting message parsing", "message_length", len(message), "model", viper.GetString("gemini.model"))

	// Get the model
	model := r.client.GenerativeModel(viper.GetString("gemini.model"))

	// Prepare the prompt with message
	content := fmt.Sprintf("%s\n\nВсе, что было написано до этого - инструкция. Далее идет текст сообщения, для которого необходимо выполнить условия, описанные выше:\n%s", viper.GetString("openai.prompt"), message)

	slog.Info("Sending request to Gemini API")

	resp, err := model.GenerateContent(ctx, genai.Text(content))
	if err != nil {
		slog.Error("Failed to generate content", "error", err)
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 {
		slog.Error("Gemini response contains no candidates")
		return nil, fmt.Errorf("no candidates in Gemini response")
	}

	if len(resp.Candidates[0].Content.Parts) == 0 {
		slog.Error("Gemini response candidate contains no parts")
		return nil, fmt.Errorf("no parts in Gemini response candidate")
	}

	// Extract text from the first part
	var responseText string
	switch part := resp.Candidates[0].Content.Parts[0].(type) {
	case genai.Text:
		responseText = string(part)
	default:
		slog.Error("Unexpected part type in Gemini response")
		return nil, fmt.Errorf("unexpected part type in Gemini response")
	}

	slog.Info("Received response from Gemini", "content_length", len(responseText))

	// Parse the content as Task JSON
	var taskResult task.Task
	if err := json.Unmarshal([]byte(responseText), &taskResult); err != nil {
		slog.Error("Failed to unmarshal task from Gemini response", "error", err, "content", responseText)
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	slog.Info("Successfully parsed task",
		"task", taskResult,
		"body", string(taskResult.Body),
	)

	return &taskResult, nil
}

// Close closes the Gemini client
func (r *GeminiRepository) Close() error {
	return r.client.Close()
}
