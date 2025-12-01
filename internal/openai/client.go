package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"auto-git/internal/provider"
)

const (
	DefaultOpenAIBaseURL    = "https://api.openai.com/v1"
	DefaultSiliconFlowURL   = "https://api.siliconflow.cn/v1"
	DefaultTimeout          = 60 * time.Second
	EnvOpenAIAPIKey         = "OPENAI_API_KEY"
	EnvSiliconFlowAPIKey    = "SILICON_KEY"
)

type Client struct {
	BaseURL string
	Client  *http.Client
	APIKey  string
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type ModelsResponse struct {
	Data []struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}

func NewClient(baseURL, apiKey string, isSiliconFlow bool) *Client {
	if baseURL == "" {
		if isSiliconFlow {
			baseURL = DefaultSiliconFlowURL
		} else {
			baseURL = DefaultOpenAIBaseURL
		}
	}

	if apiKey == "" {
		if isSiliconFlow {
			apiKey = strings.TrimSpace(getEnv(EnvSiliconFlowAPIKey))
		} else {
			apiKey = strings.TrimSpace(getEnv(EnvOpenAIAPIKey))
		}
	}

	return &Client{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: DefaultTimeout,
		},
		APIKey: strings.TrimSpace(apiKey),
	}
}

func (c *Client) ListModels() ([]provider.Model, error) {
	url := fmt.Sprintf("%s/models", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.attachAuth(req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var modelsResp ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]provider.Model, 0, len(modelsResp.Data))
	for _, m := range modelsResp.Data {
		models = append(models, provider.Model{
			Name: m.ID,
		})
	}

	return models, nil
}

func (c *Client) GenerateCommitMessage(model string, systemPrompt, userPrompt string) (string, error) {
	url := fmt.Sprintf("%s/chat/completions", c.BaseURL)

	messages := []ChatMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userPrompt,
		},
	}

	reqBody := ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	c.attachAuth(req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 || chatResp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("empty response from model")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func (c *Client) CheckConnection() error {
	// Try to list models as a connection check
	url := fmt.Sprintf("%s/models", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	c.attachAuth(req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to API server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) attachAuth(req *http.Request) {
	if c.APIKey == "" {
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
}

// getEnv gets environment variable value
func getEnv(key string) string {
	return os.Getenv(key)
}

