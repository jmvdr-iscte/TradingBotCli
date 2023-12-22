package open_ai

import (
	"github.com/jmvdr-iscte/TradingBot/initialize"
	"github.com/sashabaranov/go-openai"
)

const Prompt = "Answer only with whole numbers.Rate from 1-100 the impact that this headline has on the company.Headline:"

type OpenAIConfig struct {
	OpenAIKey string
}

func GetClient() *openai.Client {

	cfg := initialize.LoadOpenAIClient()
	return openai.NewClient(cfg.OpenAIKey)
}
