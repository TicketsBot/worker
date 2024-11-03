package event

import (
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/rpc"
	"github.com/rxdn/gdl/cache"
	"go.uber.org/zap"
)

type KafkaConsumer struct {
	logger *zap.Logger
	cache  *cache.PgCache
}

var _ rpc.Listener = (*KafkaConsumer)(nil)

func NewKafkaListener(logger *zap.Logger, cache *cache.PgCache) *KafkaConsumer {
	return &KafkaConsumer{
		logger: logger,
		cache:  cache,
	}
}

func (k *KafkaConsumer) HandleMessage(message []byte) {
	var event eventforwarding.Event
	if err := json.Unmarshal(message, &event); err != nil {
		k.logger.Error("Failed to unmarshal event", zap.Error(err))
		return
	}

	workerCtx := &worker.Context{
		Token:        event.BotToken,
		BotId:        event.BotId,
		IsWhitelabel: event.IsWhitelabel,
		ShardId:      event.ShardId,
		Cache:        k.cache,
		RateLimiter:  nil, // Use http-proxy ratelimit functionality
	}

	if err := execute(workerCtx, event.Event); err != nil {
		k.logger.Error("Failed to handle event", zap.Error(err))
	}
}
