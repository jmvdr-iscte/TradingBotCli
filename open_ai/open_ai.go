// Package open_ai connects to the openAI SDK.
package open_ai

import (
	"github.com/jmvdr-iscte/TradingBotCli/initialize"
	"github.com/sashabaranov/go-openai"
)

// The prompt to call openAI.
const Prompt = "Answer only with whole numbers.Rate from 1-100 the impact that this headline has on the company, anything above 50 is considered positive impact and below 50 is considered negative impact.Headline:"

// OpenAIConfig uses the OpenAIKey.
type OpenAIConfig struct {
	OpenAIKey string
}

// GetClient returns an openAI client.
func GetClient() *openai.Client {

	cfg := initialize.LoadOpenAIClient()
	return openai.NewClient(cfg.OpenAIKey)
}
