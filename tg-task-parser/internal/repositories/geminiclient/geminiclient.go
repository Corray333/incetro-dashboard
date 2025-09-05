package geminiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/corray333/tg-task-parser/internal/entities/task"
	"github.com/spf13/viper"
)

type GeminiRepository struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

// NewGeminiRepository creates a new Gemini repository with configuration from environment variables and viper
func NewGeminiRepository() (*GeminiRepository, error) {
	// Get API key from environment variable (sensitive data)
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		slog.Error("GEMINI_API_KEY environment variable is not set")
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}

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

	slog.Info("Gemini repository initialized", "model", viper.GetString("gemini.model"), "proxy_enabled", proxyURL != "")

	return &GeminiRepository{
		client:  httpClient,
		apiKey:  apiKey,
		baseURL: "https://generativelanguage.googleapis.com/v1beta",
	}, nil
}

// GeminiRequest represents the request structure for Gemini API
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiResponse represents the response structure from Gemini API
type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

// ParseMessage sends a message to Gemini API and parses the response into a Task struct
func (r *GeminiRepository) ParseMessage(ctx context.Context, message string) (*task.Task, error) {
	// Validate input message
	if strings.TrimSpace(message) == "" {
		slog.Error("Empty message provided for parsing")
		return nil, fmt.Errorf("message cannot be empty")
	}

	slog.Info("Starting message parsing", "message_length", len(message), "model", viper.GetString("gemini.model"))

	// Prepare the prompt with message
	content := fmt.Sprintf("%s\n\nВсе, что было написано до этого - инструкция. Далее идет текст сообщения, для которого необходимо выполнить условия, описанные выше:\n%s", viper.GetString("openai.prompt"), message)

	// Prepare request body
	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{
						Text: content,
					},
				},
			},
		},
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		slog.Error("Failed to marshal request body", "error", err)
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Build URL
	model := viper.GetString("gemini.model")
	if model == "" {
		model = "gemini-1.5-flash"
	}
	url := fmt.Sprintf("%s/models/%s:generateContent", r.baseURL, model)

	slog.Info("Sending request to Gemini API", "url", url)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		slog.Error("Failed to create HTTP request", "error", err)
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-goog-api-key", r.apiKey)

	// Send request
	resp, err := r.client.Do(req)
	if err != nil {
		slog.Error("Failed to send HTTP request", "error", err)
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read response body", "error", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("Gemini API returned error", "status_code", resp.StatusCode, "response", string(respBodyBytes))
		return nil, fmt.Errorf("Gemini API returned status %d: %s", resp.StatusCode, string(respBodyBytes))
	}

	// Parse Gemini response
	var geminiResp GeminiResponse
	if err := json.Unmarshal(respBodyBytes, &geminiResp); err != nil {
		slog.Error("Failed to unmarshal Gemini response", "error", err, "response", string(respBodyBytes))
		return nil, fmt.Errorf("failed to unmarshal Gemini response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 {
		slog.Error("Gemini response contains no candidates")
		return nil, fmt.Errorf("no candidates in Gemini response")
	}

	if len(geminiResp.Candidates[0].Content.Parts) == 0 {
		slog.Error("Gemini response candidate contains no parts")
		return nil, fmt.Errorf("no parts in Gemini response candidate")
	}

	responseText := geminiResp.Candidates[0].Content.Parts[0].Text

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
	// HTTP client doesn't need explicit closing
	return nil
}
