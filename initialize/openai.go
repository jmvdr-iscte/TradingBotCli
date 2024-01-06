// Package initialize serves to initialize the configs
package initialize

import (
	"os"
)

const Prompt = "Answer only with whole numbers.Rate from 1-100 the impact that this headline has on the company.Headline:"

// OpenAIConfig is the initial OpenAi config.
type OpenAIConfig struct {
	OpenAIKey string
}

// LoadOpenAIClient loads the initial config with the .env values.
func LoadOpenAIClient() *OpenAIConfig {
	cfg := &OpenAIConfig{
		OpenAIKey: "",
	}

	if open_ai_key, exists := os.LookupEnv("OPEN_AI_KEY"); exists {
		cfg.OpenAIKey = open_ai_key
	}

	return cfg
}
