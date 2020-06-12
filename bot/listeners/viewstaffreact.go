package listeners

import (
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"github.com/rxdn/gdl/rest"
)

func OnViewStaffReact(worker *worker.Context, e *events.MessageReactionAdd, extra eventforwarding.Extra) {
	// Create error context for later
	errorContext := errorcontext.WorkerErrorContext{
		Guild:   e.GuildId,
		User:    e.UserId,
		Channel: e.ChannelId,
		Shard:   worker.ShardId,
		Command: "viewstaff",
	}

	// In DMs
	if e.GuildId == 0 {
		return
	}

	// ignore self
	if e.UserId == worker.BotId {
		return
	}

	page, isViewStaffMessage := redis.GetPage(redis.Client, e.MessageId)
	if !isViewStaffMessage {
		return
	}

	if e.Emoji.Name == "▶️" {
		page++
	} else if e.Emoji.Name == "◀️" {
		page--
	} else {
		return
	}

	_ = worker.DeleteUserReaction(e.ChannelId, e.MessageId, e.UserId, e.Emoji.Name) // TODO: Permission check

	if page < 0 {
		return
	}

	_, err := worker.EditMessage(e.ChannelId, e.MessageId, rest.EditMessageData{
		Embed: logic.BuildViewStaffMessage(e.GuildId, worker, page, errorContext),
	})
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	// TODO: Race condition? I'm too tired rn
	redis.SetPage(redis.Client, e.MessageId, page)
}
