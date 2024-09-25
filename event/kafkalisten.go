package event

import (
	"context"
	"errors"
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/metrics/prometheus"
	"github.com/TicketsBot/worker/config"
	"github.com/rxdn/gdl/cache"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type KafkaConsumer struct {
	client  *kgo.Client
	logger  *zap.Logger
	cache   *cache.PgCache
	running *atomic.Bool

	cancelFunc context.CancelFunc
}

const kafkaConsumerGroup = "worker"
const maxEventsPerPoll = 100

func ConnectKafka(logger *zap.Logger, cache *cache.PgCache) (*KafkaConsumer, error) {
	kafkaClient, err := connectKafka()
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		client:  kafkaClient,
		logger:  logger,
		cache:   cache,
		running: atomic.NewBool(false),
	}, nil
}

func (k *KafkaConsumer) Run() {
	if k.running.Swap(true) {
		k.logger.Fatal("Kafka client already running")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	k.cancelFunc = cancel

	for {
		select {
		case <-ctx.Done():
			return
		default:
			records, err := k.poll(ctx)
			if err != nil {
				if errors.Is(err, kgo.ErrClientClosed) {
					k.logger.Info("Kafka client closed, stopping read loop")
					return
				} else if errors.Is(err, context.Canceled) {
					k.logger.Info("Context cancelled, stopping read loop")
					return
				} else {
					k.logger.Error("Failed to poll records", zap.Error(err))
					continue
				}
			}

			for _, record := range records {
				record := record

				go func() {
					var event eventforwarding.Event
					if err := json.Unmarshal(record, &event); err != nil {
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
				}()
			}
		}
	}
}

func (k *KafkaConsumer) Shutdown() {
	k.client.Close()
	k.cancelFunc()
}

func (k *KafkaConsumer) poll(ctx context.Context) ([][]byte, error) {
	fetches := k.client.PollRecords(ctx, maxEventsPerPoll)
	if fetches.IsClientClosed() {
		return nil, kgo.ErrClientClosed
	}

	if err := fetches.Err(); err != nil {
		return nil, err
	}

	records := make([][]byte, 0, fetches.NumRecords())
	prometheus.KafkaBatchSize.Observe(float64(fetches.NumRecords()))

	iter := fetches.RecordIter()
	for !iter.Done() {
		record := iter.Next()
		records = append(records, record.Value)
	}

	return records, nil
}

func connectKafka() (*kgo.Client, error) {
	return kgo.NewClient(
		kgo.SeedBrokers(config.Conf.Kafka.Brokers...),
		kgo.ConsumerGroup(kafkaConsumerGroup),
		kgo.ConsumeTopics(config.Conf.Kafka.EventsTopic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()),
	)
}
