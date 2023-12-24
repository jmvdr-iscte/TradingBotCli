package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hibiken/asynq"
	"github.com/jmvdr-iscte/TradingBotCli/models"
	"github.com/jmvdr-iscte/TradingBotCli/open_ai"
	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
)

const TaskProcessOrder = "task:process_order"

// TODO Replace every print with logs
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

func (processor *RedisTaskProcessor) ProcessTaskProcessOrder(ctx context.Context, task *asynq.Task) error {
	var payload models.Message

	if err := json.Unmarshal(task.Payload(), &payload); err != nil { // guarda na referencia da memória da variavel

		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}
	log.Info().Msgf("Processing task: %v", task.ResultWriter().TaskID())

	response, err := askGPT(processor.openai_client, payload)
	if err != nil {
		return fmt.Errorf("failed asking chat gpt: %w", asynq.SkipRetry)
	}

	if response >= 75 {
		if err := processor.alpaca_client.BuyPosition(response, payload.Symbols[0], payload.Risk); err != nil {
			return fmt.Errorf("failed to buy: %w", asynq.SkipRetry)
		}
		fmt.Println("Buy: ", payload)
		return nil

	} else if response <= 25 && response > 0 {

		if err := processor.alpaca_client.SellPosition(payload.Symbols[0], response, payload.Risk); err != nil {
			return fmt.Errorf("failed to sell, or short: %w", err)
		}
		fmt.Println("Sell: ", payload)
		return nil
	}
	return nil
}

func askGPT(client *openai.Client, m models.Message) (int, error) {
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
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
