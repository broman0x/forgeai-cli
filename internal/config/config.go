package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Provider string
	Model    string
	FirstRun bool
}

func Load() *Config {
	return &Config{
		Provider: viper.GetString("provider"),
		Model:    viper.GetString("model"),
		FirstRun: viper.GetBool("first_run"),
	}
}

func SaveFirstRun(status bool) {
	viper.Set("first_run", status)
	viper.WriteConfig()
}

func CreateEnvFile(apiKey string) error {
	content := fmt.Sprintf(`GEMINI_API_KEY=%s
GEMINI_MODEL=gemini-2.5-flash
# OPENROUTER_API_KEY=
`, apiKey)

	return os.WriteFile(".env", []byte(content), 0644)
}