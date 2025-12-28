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

type OpenAIProvider struct {
	ApiKey  string
	Model   string
	Client  *http.Client
	History []openAIMessage
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func newOpenAIProvider(apiKey, model string) *OpenAIProvider {
	return &OpenAIProvider{
		ApiKey:  apiKey,
		Model:   model,
		Client:  &http.Client{Timeout: 120 * time.Second},
		History: []openAIMessage{},
	}
}

func (o *OpenAIProvider) Name() string {
	return "OpenAI (" + o.Model + ")"
}

func (o *OpenAIProvider) Reset() {
	o.History = []openAIMessage{}
}

func (o *OpenAIProvider) Send(prompt string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	userMsg := openAIMessage{Role: "user", Content: prompt}
	currentContext := append(o.History, userMsg)

	payload, _ := json.Marshal(openAIRequest{
		Model:    o.Model,
		Messages: currentContext,
	})

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.ApiKey)

	resp, err := o.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var res openAIResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return "", fmt.Errorf("parse error: %s", string(body))
	}

	if res.Error != nil {
		return "", fmt.Errorf("openai error: %s", res.Error.Message)
	}

	if len(res.Choices) > 0 {
		ans := strings.TrimSpace(res.Choices[0].Message.Content)
		o.History = append(currentContext, openAIMessage{Role: "assistant", Content: ans})
		return ans, nil
	}

	return "", fmt.Errorf("empty response")
}
