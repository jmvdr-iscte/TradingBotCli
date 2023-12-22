package worker

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/jmvdr-iscte/TradingBot/models"
)

type TaskDistributor interface {
	DistributeTaskProcessOrder(
		ctx context.Context,
		order *models.Message,
		opts ...asynq.Option,
	) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor { // estamos a forçar a struct a implementar a interface
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}
