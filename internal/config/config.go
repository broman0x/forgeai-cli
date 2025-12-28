package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Language     string `json:"language"`
	FirstRun     bool   `json:"first_run"`
	LastModel    string `json:"last_model"`
	LastProvider string `json:"last_provider"`
	InstallPath  string `json:"install_path"`
	Version      string `json:"version"`
}

var globalConfig *Config

func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	if os.Getenv("APPDATA") != "" {
		return filepath.Join(os.Getenv("APPDATA"), "ForgeAI", "config.json")
	}
	return filepath.Join(home, ".config", "forgeai", "config.json")
}

func Load() *Config {
	if globalConfig != nil {
		return globalConfig
	}

	configPath := GetConfigPath()
	data, err := os.ReadFile(configPath)

	if err != nil {
		globalConfig = &Config{
			Language: "en",
			FirstRun: true,
			Version:  "1.0.1",
		}
		return globalConfig
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		globalConfig = &Config{
			Language: "en",
			FirstRun: true,
			Version:  "1.0.1",
		}
		return globalConfig
	}

	if cfg.Version == "" {
		cfg.FirstRun = true
	}

	globalConfig = &cfg
	return globalConfig
}

func ResetCache() {
	globalConfig = nil
}

func Save(cfg *Config) error {
	configPath := GetConfigPath()
	dir := filepath.Dir(configPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func SaveLastModel(provider, model string) error {
	cfg := Load()
	cfg.LastProvider = provider
	cfg.LastModel = model
	return Save(cfg)
}

func SaveLanguage(language string) error {
	cfg := Load()
	cfg.Language = language
	return Save(cfg)
}

func SaveFirstRun(status bool) error {
	cfg := Load()
	cfg.FirstRun = status
	return Save(cfg)
}

func CreateEnvFile(apiKey string) error {
	content := fmt.Sprintf("GEMINI_API_KEY=%s\n", apiKey)
	return os.WriteFile(".env", []byte(content), 0644)
}

func SaveAPIKey(keyName, keyValue string) error {
	envPath := ".env"

	existing, err := os.ReadFile(envPath)
	envContent := ""
	if err == nil {
		envContent = string(existing)
	}

	lines := strings.Split(envContent, "\n")
	found := false
	var newLines []string

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), keyName+"=") {
			newLines = append(newLines, fmt.Sprintf("%s=%s", keyName, keyValue))
			found = true
		} else if strings.TrimSpace(line) != "" {
			newLines = append(newLines, line)
		}
	}

	if !found {
		newLines = append(newLines, fmt.Sprintf("%s=%s", keyName, keyValue))
	}

	content := strings.Join(newLines, "\n") + "\n"
	return os.WriteFile(envPath, []byte(content), 0644)
}
