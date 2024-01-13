// Package worker encapsules all the asynq modules.
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hibiken/asynq"
	"github.com/jmvdr-iscte/TradingBotCli/enums"
	"github.com/jmvdr-iscte/TradingBotCli/models"
	"github.com/jmvdr-iscte/TradingBotCli/open_ai"
	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
)

const TaskProcessOrder = "task:process_order"

// DistributeTaskProcessOrder returns an error if anything goes wrong with
// distributing the tasks to a redis queue. If it was able to distribute it
// it returns nil.
func (distributor *RedisTaskDistributor) DistributeTaskProcessOrder(
	ctx context.Context,
	order *models.Message,
	opts ...asynq.Option,
) error {
	json_payload, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload %w", err)
	}

	task := asynq.NewTask(TaskProcessOrder, json_payload, opts...)

	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task %w", err)
	}
	log.Info().Msgf("The details are %v ", info)
	return nil
}

// ProcessTaskProcessOrder returns an error if it was not able to process the task.
// It is responsible for the sentiment analysis and caling the alpaca sdk in order to
// sell or buy.
func (processor *RedisTaskProcessor) ProcessTaskProcessOrder(ctx context.Context, task *asynq.Task) error {
	riskLevels := map[enums.Risk]bool{
		enums.Power: true,
		enums.Safe:  true,
	}

	var high_limit = 75
	var low_limit = 25
	var payload models.Message

	if err := json.Unmarshal(task.Payload(), &payload); err != nil { // guarda na referencia da memória da variavel

		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}
	log.Info().Msgf("Processing task: %v", task.ResultWriter().TaskID())

	response, err := sentimentAnalysis(processor.openai_client, payload)
	if err != nil {
		return fmt.Errorf("failed asking chat gpt: %w", asynq.SkipRetry)
	}

	if riskLevels[payload.Risk] {
		high_limit = 95
		low_limit = 5
	}

	if response >= high_limit {
		if err := processor.alpaca_client.BuyPosition(response, payload.Symbols[0], payload.Risk); err != nil {
			return fmt.Errorf("failed to buy: %w", asynq.SkipRetry)
		}
		fmt.Println("Buy: ", payload)
		return nil

	} else if response <= low_limit && response > 0 {

		if err := processor.alpaca_client.SellPosition(payload.Symbols[0], response, payload.Risk); err != nil {
			return fmt.Errorf("failed to sell, or short: %w", err)
		}
		fmt.Println("Sell: ", payload)
		return nil
	}
	return nil
}

// sentimentAnalysis calls the openAI sdk in order to get a sentiment analysis given a certain stock
// it returns a response that matches the sentiment analysis. Also it returns an error if
// it's not able to correctly process the input.
func sentimentAnalysis(client *openai.Client, m models.Message) (int, error) {
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4TurboPreview,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: open_ai.Prompt + m.Headline,
				},
			},
		},
	)

	if err != nil {
		return 0, err
	}

	result, err := strconv.Atoi(strings.TrimSpace(resp.Choices[0].Message.Content))
	fmt.Println("The sentiment analysis is :", resp.Choices[0].Message.Content)
	if err != nil {
		return 0, fmt.Errorf("conversion error: %v", err)
	}
	return result, nil
}
