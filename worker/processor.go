package worker

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"

	"github.com/jmvdr-iscte/TradingBotCli/alpaca"
	"github.com/jmvdr-iscte/TradingBotCli/open_ai"
	"github.com/sashabaranov/go-openai"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskProcessOrder(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server        *asynq.Server
	alpaca_client *alpaca.AlpacaClient
	openai_client *openai.Client
}

func NewRedisTaskProcessor(
	redisOpt asynq.RedisClientOpt,
) TaskProcessor {
	//Add list priorities
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  6,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().Err(err).Str("type", task.Type()).
					Bytes("payload", task.Payload()).Msg("process task failed")
			}),
		},
	)
	alpaca_client := alpaca.LoadClient()
	openai_client := open_ai.GetClient()
	return &RedisTaskProcessor{
		server:        server,
		alpaca_client: alpaca_client,
		openai_client: openai_client,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux() //register each task
	mux.HandleFunc(TaskProcessOrder, processor.ProcessTaskProcessOrder)
	return processor.server.Start(mux)
}
