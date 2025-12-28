package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ClaudeProvider struct {
	ApiKey  string
	Model   string
	Client  *http.Client
	History []claudeMessage
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeRequest struct {
	Model     string          `json:"model"`
	Messages  []claudeMessage `json:"messages"`
	MaxTokens int             `json:"max_tokens"`
}

type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func newClaudeProvider(apiKey, model string) *ClaudeProvider {
	return &ClaudeProvider{
		ApiKey:  apiKey,
		Model:   model,
		Client:  &http.Client{Timeout: 120 * time.Second},
		History: []claudeMessage{},
	}
}

func (c *ClaudeProvider) Name() string {
	return "Claude (" + c.Model + ")"
}

func (c *ClaudeProvider) Reset() {
	c.History = []claudeMessage{}
}

func (c *ClaudeProvider) Send(prompt string) (string, error) {
	url := "https://api.anthropic.com/v1/messages"

	userMsg := claudeMessage{Role: "user", Content: prompt}
	currentContext := append(c.History, userMsg)

	payload, _ := json.Marshal(claudeRequest{
		Model:     c.Model,
		Messages:  currentContext,
		MaxTokens: 4096,
	})

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.ApiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var res claudeResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return "", fmt.Errorf("parse error: %s", string(body))
	}

	if res.Error != nil {
		return "", fmt.Errorf("claude error: %s", res.Error.Message)
	}

	if len(res.Content) > 0 {
		ans := strings.TrimSpace(res.Content[0].Text)
		c.History = append(currentContext, claudeMessage{Role: "assistant", Content: ans})
		return ans, nil
	}

	return "", fmt.Errorf("empty response")
}
