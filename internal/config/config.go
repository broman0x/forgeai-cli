package config

import (
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
