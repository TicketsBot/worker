package rpc

import (
	"context"
	"errors"
	"github.com/TicketsBot/worker/bot/metrics/prometheus"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/config"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type Client struct {
	client  *kgo.Client
	logger  *zap.Logger
	running *atomic.Bool

	listeners map[string]Listener

	cancelFunc context.CancelFunc
}

const kafkaConsumerGroup = "worker"
const maxEventsPerPoll = 100

func NewClient(logger *zap.Logger, listeners map[string]Listener) (*Client, error) {
	kafkaClient, err := connectKafka(utils.Keys(listeners))
	if err != nil {
		return nil, err
	}

	return &Client{
		client:    kafkaClient,
		logger:    logger,
		running:   atomic.NewBool(false),
		listeners: listeners,
	}, nil
}

func (k *Client) Run() {
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
				listener, ok := k.listeners[record.Topic]
				if !ok {
					k.logger.Warn("No listener found for topic", zap.String("topic", record.Topic))
					continue
				}

				value := record.Value
				go listener.HandleMessage(value)
			}
		}
	}
}

func (k *Client) Shutdown() {
	k.client.Close()
	k.cancelFunc()
}

func (k *Client) poll(ctx context.Context) ([]*kgo.Record, error) {
	fetches := k.client.PollRecords(ctx, maxEventsPerPoll)
	if fetches.IsClientClosed() {
		return nil, kgo.ErrClientClosed
	}

	if err := fetches.Err(); err != nil {
		return nil, err
	}

	records := make([]*kgo.Record, 0, fetches.NumRecords())
	prometheus.KafkaBatchSize.Observe(float64(fetches.NumRecords()))

	iter := fetches.RecordIter()
	for !iter.Done() {
		record := iter.Next()
		records = append(records, record)
	}

	return records, nil
}

func connectKafka(topics []string) (*kgo.Client, error) {
	return kgo.NewClient(
		kgo.SeedBrokers(config.Conf.Kafka.Brokers...),
		kgo.ConsumerGroup(kafkaConsumerGroup),
		kgo.ConsumeTopics(topics...),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()),
	)
}
