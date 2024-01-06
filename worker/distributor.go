// Package worker encapsules all the asynq modules.
package worker

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/jmvdr-iscte/TradingBotCli/models"
)

// TaskDistributor interface has all the tasks to run.
type TaskDistributor interface {
	DistributeTaskProcessOrder(
		ctx context.Context,
		order *models.Message,
		opts ...asynq.Option,
	) error
}

// RedisTaskDistributor is the asynq client.
type RedisTaskDistributor struct {
	client *asynq.Client
}

// NewRedisTaskDistributor returns a new TaskDistributor with the given opts
func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor { // estamos a forçar a struct a implementar a interface
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}
