package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/broman0x/forgeai-cli/internal/config"
	"github.com/spf13/viper"
)

type Provider interface {
	Send(prompt string) (string, error)
	Name() string
	Reset()
}

func CreateProvider(pType, modelName string) (Provider, error) {
	pType = strings.ToLower(pType)

	switch pType {
	case "gemini":
		key := os.Getenv("GEMINI_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY not found in .env")
		}
		if modelName == "" {
			modelName = "gemini-2.5-flash"
		}
		return newGeminiProvider(key, modelName), nil

	case "openai", "chatgpt":
		key := os.Getenv("OPENAI_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY not found in .env")
		}
		if modelName == "" {
			modelName = "gpt-3.5-turbo"
		}
		return newOpenAIProvider(key, modelName), nil

	case "claude", "anthropic":
		key := os.Getenv("ANTHROPIC_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY not found in .env")
		}
		if modelName == "" {
			modelName = "claude-3-haiku-20240307"
		}
		return newClaudeProvider(key, modelName), nil

	case "ollama":
		if !isOllamaRunning() {
			return nil, fmt.Errorf("ollama is not running")
		}
		if modelName == "" {
			modelName = "llama3"
		}
		return newOllamaProvider(modelName), nil

	default:
		return nil, fmt.Errorf("unknown provider type: %s", pType)
	}
}

func NewProvider() (Provider, error) {
	cfg := config.Load()

	if cfg.LastProvider != "" && cfg.LastModel != "" {
		prov, err := CreateProvider(cfg.LastProvider, cfg.LastModel)
		if err == nil {
			return prov, nil
		}
	}

	if isOllamaRunning() {
		return CreateProvider("ollama", "llama3")
	}

	configProvider := strings.ToLower(viper.GetString("provider"))
	if configProvider != "" {
		prov, err := CreateProvider(configProvider, "")
		if err == nil {
			return prov, nil
		}
	}

	if os.Getenv("GEMINI_API_KEY") != "" {
		return CreateProvider("gemini", "gemini-2.5-flash")
	}

	if os.Getenv("OPENAI_API_KEY") != "" {
		return CreateProvider("openai", "gpt-3.5-turbo")
	}

	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return CreateProvider("claude", "claude-3-haiku-20240307")
	}

	return nil, fmt.Errorf("no AI provider available. Please set up API key or start Ollama")
}

func getOllamaHost() string {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		host = "localhost"
	}
	return host
}

func getOllamaPort() string {
	port := os.Getenv("OLLAMA_PORT")
	if port == "" {
		port = "11434"
	}
	return port
}

func isOllamaRunning() bool {
	host := getOllamaHost()
	port := getOllamaPort()
	addr := fmt.Sprintf("%s:%s", host, port)

	conn, err := net.DialTimeout("tcp", addr, 1000*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

type GeminiProvider struct {
	ApiKey  string
	Model   string
	Client  *http.Client
	History []geminiContent
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}
type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}
type geminiPart struct {
	Text string `json:"text"`
}
type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func newGeminiProvider(apiKey, model string) *GeminiProvider {
	return &GeminiProvider{
		ApiKey:  apiKey,
		Model:   model,
		Client:  &http.Client{Timeout: 120 * time.Second},
		History: []geminiContent{},
	}
}

func (g *GeminiProvider) Name() string { return "Gemini (" + g.Model + ")" }
func (g *GeminiProvider) Reset()       { g.History = []geminiContent{} }

func (g *GeminiProvider) Send(prompt string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", g.Model, g.ApiKey)

	userMsg := geminiContent{Role: "user", Parts: []geminiPart{{Text: prompt}}}
	currentContext := append(g.History, userMsg)

	payload, _ := json.Marshal(geminiRequest{Contents: currentContext})

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var res geminiResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return "", fmt.Errorf("parse error: %s", string(body))
	}

	if res.Error != nil {
		return "", fmt.Errorf("gemini error (%d): %s", res.Error.Code, res.Error.Message)
	}

	if len(res.Candidates) > 0 && len(res.Candidates[0].Content.Parts) > 0 {
		ans := strings.TrimSpace(res.Candidates[0].Content.Parts[0].Text)
		g.History = append(currentContext, geminiContent{Role: "model", Parts: []geminiPart{{Text: ans}}})
		return ans, nil
	}

	return "", fmt.Errorf("empty response")
}

type OllamaProvider struct {
	BaseURL string
	Model   string
	Client  *http.Client
	History []ollamaMessage
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaResponse struct {
	Message ollamaMessage `json:"message"`
	Done    bool          `json:"done"`
}

func newOllamaProvider(model string) *OllamaProvider {
	host := getOllamaHost()
	port := getOllamaPort()
	baseURL := fmt.Sprintf("http://%s:%s/api/chat", host, port)

	return &OllamaProvider{
		BaseURL: baseURL,
		Model:   model,
		Client:  &http.Client{Timeout: 300 * time.Second},
		History: []ollamaMessage{},
	}
}

func (o *OllamaProvider) Name() string { return "Ollama (" + o.Model + ")" }
func (o *OllamaProvider) Reset()       { o.History = []ollamaMessage{} }

func (o *OllamaProvider) Send(prompt string) (string, error) {
	o.History = append(o.History, ollamaMessage{Role: "user", Content: prompt})

	payload, _ := json.Marshal(ollamaRequest{
		Model:    o.Model,
		Messages: o.History,
		Stream:   false,
	})

	req, _ := http.NewRequest("POST", o.BaseURL, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.Client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "deadline exceeded") {
			return "", fmt.Errorf("timeout: model took too long to respond. Try a smaller file or faster model")
		}
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama error %d: %s", resp.StatusCode, string(body))
	}

	var res ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("decode error: %v", err)
	}

	ans := strings.TrimSpace(res.Message.Content)
	o.History = append(o.History, res.Message)
	return ans, nil
}
