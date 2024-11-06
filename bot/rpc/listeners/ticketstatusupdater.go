package listeners

import (
	"context"
	"encoding/json"
	"github.com/TicketsBot/common/rpc"
	"github.com/TicketsBot/common/rpc/model"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/metrics/prometheus"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/rxdn/gdl/cache"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/rest"
	"go.uber.org/zap"
)

type TicketStatusUpdater struct {
	*BaseListener
	logger *zap.Logger
}

var _ rpc.Listener = (*TicketStatusUpdater)(nil)

func NewTicketStatusUpdater(cache *cache.PgCache, logger *zap.Logger) *TicketStatusUpdater {
	return &TicketStatusUpdater{
		BaseListener: NewBaseListener(cache),
		logger:       logger,
	}
}

func (u *TicketStatusUpdater) HandleMessage(ctx context.Context, message []byte) {
	var event model.TicketStatusUpdate
	if err := json.Unmarshal(message, &event); err != nil {
		u.logger.Error("Failed to unmarshal event", zap.Error(err))
		return
	}

	worker, err := u.ContextForGuild(ctx, event.GuildId)
	if err != nil {
		u.logger.Error("Failed to get worker context", zap.Error(err))
		return
	}

	canMove, err := u.CategoryHasSpace(ctx, worker, event)
	if err != nil {
		u.logger.Error("Failed to check category space", zap.Error(err))
		return
	}

	if !canMove {
		u.logger.Debug(
			"Tried to move ticket to updated status category, but it has no space",
			zap.Uint64("guild_id", event.GuildId),
			zap.Uint64("category_id", event.NewCategoryId),
		)
		return
	}

	// Don't move the ticket if it's already in the correct category
	ch, err := worker.GetChannel(event.ChannelId)
	if err != nil {
		u.logger.Error(
			"Failed to get ticket channel",
			zap.Error(err),
			zap.Uint64("channel_id", event.ChannelId),
			zap.Uint64("guild_id", event.GuildId),
		)
		return
	}

	if ch.ParentId.Value == event.NewCategoryId {
		u.logger.Debug(
			"Ticket is already in the correct category",
			zap.Uint64("channel_id", event.ChannelId),
			zap.Uint64("category_id", event.NewCategoryId),
		)
		return
	}

	if _, err := worker.ModifyChannel(event.ChannelId, rest.ModifyChannelData{
		ParentId: event.NewCategoryId,
	}); err != nil {
		u.logger.Error(
			"Failed to move ticket to updated status category",
			zap.Error(err),
			zap.Uint64("channel_id", event.ChannelId),
			zap.Uint64("guild_id", event.GuildId),
			zap.Uint64("category_id", event.NewCategoryId),
		)
		return
	}

	prometheus.CategoryUpdates.Inc()
	u.logger.Debug("Moved ticket to updated status category", zap.Uint64("channel_id", event.ChannelId), zap.Uint64("category_id", event.NewCategoryId))
}

func (u *TicketStatusUpdater) CategoryHasSpace(ctx context.Context, worker *worker.Context, event model.TicketStatusUpdate) (bool, error) {
	channels, err := u.cache.GetGuildChannels(ctx, event.GuildId)
	if err != nil {
		return false, err
	}

	if u.countCategoryChannels(channels, event.NewCategoryId) < 50 {
		return true, nil
	}

	// Try refreshing the channels in the cache if it hasn't been done recently
	canRetry, err := redis.TakeChannelRefetchToken(ctx, event.GuildId)
	if err != nil {
		return false, err
	}

	if canRetry {
		channels, err := rest.GetGuildChannels(ctx, worker.Token, nil, event.GuildId)
		if err != nil {
			return false, err
		}

		return u.countCategoryChannels(channels, event.NewCategoryId) < 50, nil
	} else {
		return false, nil
	}
}

func (u *TicketStatusUpdater) countCategoryChannels(channels []channel.Channel, categoryId uint64) int {
	count := 0
	for _, ch := range channels {
		if ch.ParentId.Value == categoryId {
			count++
		}
	}

	return count
}
