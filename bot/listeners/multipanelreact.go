package listeners

import (
	"context"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	context2 "github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"golang.org/x/sync/errgroup"
)

func OnMultiPanelReact(worker *worker.Context, e *events.MessageReactionAdd) {
	errorContext := errorcontext.WorkerErrorContext{
		Guild:   e.GuildId,
		User:    e.UserId,
		Channel: e.ChannelId,
		Shard:   worker.ShardId,
	}

	// In DMs
	if e.GuildId == 0 {
		return
	}

	if e.UserId == worker.BotId || e.Member.User.Bot {
		return
	}

	// Get multipanel from DB
	multiPanel, ok, err := dbclient.Client.MultiPanels.GetByMessageId(e.MessageId)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	if !ok {
		return
	}

	emoji := e.Emoji.Name // This is the actual unicode emoji (https://discordapp.com/developers/docs/resources/emoji#emoji-object-gateway-reaction-standard-emoji-example)

	// get the sub-panels
	subPanels, err := dbclient.Client.MultiPanelTargets.GetPanels(multiPanel.Id)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	var panel *database.Panel

	for _, subPanel := range subPanels {
		if subPanel.ReactionEmote == emoji {
			panel = &subPanel
			break
		}
	}

	if panel == nil {
		return
	}

	// TODO: Check perms
	// Remove the reaction from the message
	if err := worker.DeleteUserReaction(e.ChannelId, e.MessageId, e.UserId, emoji); err != nil {
		sentry.LogWithContext(err, errorContext)
	}

	var blacklisted bool
	var premiumTier premium.PremiumTier

	group, _ := errgroup.WithContext(context.Background())

	// get blacklisted
	group.Go(func() (err error) {
		blacklisted, err = dbclient.Client.Blacklist.IsBlacklisted(e.GuildId, e.UserId)
		return
	})

	// get premium
	group.Go(func() (err error) {
		premiumTier, err = utils.PremiumClient.GetTierByGuildId(e.GuildId, true, worker.Token, worker.RateLimiter)
		return
	})

	if err := group.Wait(); err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	if blacklisted {
		return
	}

	panelContext := context2.NewPanelContext(worker, e.GuildId, e.ChannelId, e.UserId, premiumTier)

	go logic.OpenTicket(&panelContext, panel, panel.Title)
}
