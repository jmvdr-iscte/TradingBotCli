package initialize

import (
	"os"
)

const Prompt = "Answer only with whole numbers.Rate from 1-100 the impact that this headline has on the company.Headline:"

type OpenAIConfig struct {
	OpenAIKey string
}

func LoadOpenAIClient() *OpenAIConfig {
	cfg := &OpenAIConfig{
		OpenAIKey: "",
	}

	if open_ai_key, exists := os.LookupEnv("OPEN_AI_KEY"); exists {
		cfg.OpenAIKey = open_ai_key
	}

	return cfg
}
