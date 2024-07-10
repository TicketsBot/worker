package context

import (
	"context"
	"errors"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/rest/request"
)

type DashboardContext struct {
	context.Context
	*Replyable
	*StateCache
	worker                     *worker.Context
	guildId, channelId, userId uint64
	premium                    premium.PremiumTier
}

var _ registry.CommandContext = (*DashboardContext)(nil)

func NewDashboardContext(
	ctx context.Context,
	worker *worker.Context,
	guildId, channelId, userId uint64,
	premium premium.PremiumTier,
) DashboardContext {
	c := DashboardContext{
		Context:   ctx,
		worker:    worker,
		guildId:   guildId,
		channelId: channelId,
		userId:    userId,
		premium:   premium,
	}

	c.Replyable = NewReplyable(&c)
	c.StateCache = NewStateCache(&c)
	return c
}

func (c *DashboardContext) Worker() *worker.Context {
	return c.worker
}

func (c *DashboardContext) GuildId() uint64 {
	return c.guildId
}

func (c *DashboardContext) ChannelId() uint64 {
	return c.channelId
}

func (c *DashboardContext) UserId() uint64 {
	return c.userId
}

func (c *DashboardContext) UserPermissionLevel(ctx context.Context) (permcache.PermissionLevel, error) {
	member, err := c.Member()
	if err != nil {
		return permcache.Everyone, err
	}

	return permcache.GetPermissionLevel(ctx, utils.ToRetriever(c.worker), member, c.guildId)
}

func (c *DashboardContext) PremiumTier() premium.PremiumTier {
	return c.premium
}

func (c *DashboardContext) IsInteraction() bool {
	return true
}

func (c *DashboardContext) Source() registry.Source {
	return registry.SourceDashboard
}

func (c *DashboardContext) ToErrorContext() errorcontext.WorkerErrorContext {
	return errorcontext.WorkerErrorContext{
		Guild:   c.guildId,
		User:    c.userId,
		Channel: c.channelId,
	}
}

func (c *DashboardContext) openDm() (uint64, bool) {
	cachedId, err := redis.GetDMChannel(c.UserId(), c.Worker().BotId)
	if err != nil { // We can continue
		if err != redis.ErrNotCached {
			sentry.ErrorWithContext(err, c.ToErrorContext())
		}
	} else { // We have it cached
		if cachedId == nil {
			return 0, false
		} else {
			return *cachedId, true
		}
	}

	ch, err := c.Worker().CreateDM(c.UserId())
	if err != nil {
		// check for 403
		if err, ok := err.(request.RestError); ok && err.StatusCode == 403 {
			if err := redis.StoreNullDMChannel(c.UserId(), c.Worker().BotId); err != nil {
				sentry.ErrorWithContext(err, c.ToErrorContext())
			}

			return 0, false
		}

		sentry.ErrorWithContext(err, c.ToErrorContext())
		return 0, false
	}

	if err := redis.StoreDMChannel(c.UserId(), ch.Id, c.Worker().BotId); err != nil {
		sentry.ErrorWithContext(err, c.ToErrorContext())
	}

	return ch.Id, true
}

func (c *DashboardContext) ReplyWith(response command.MessageResponse) (message.Message, error) {
	channelId, ok := c.openDm()
	if !ok { // Error handled in openDm function
		return message.Message{}, errors.New("failed to open dm")
	}

	msg, err := c.Worker().CreateMessageComplex(channelId, response.IntoCreateMessageData())
	if err != nil {
		sentry.ErrorWithContext(err, c.ToErrorContext())
	}

	return msg, err
}

func (c *DashboardContext) Channel() (channel.PartialChannel, error) {
	ch, err := c.Worker().GetChannel(c.channelId)
	if err != nil {
		return channel.PartialChannel{}, err
	}

	return ch.ToPartialChannel(), nil
}

func (c *DashboardContext) Guild() (guild.Guild, error) {
	return c.Worker().GetGuild(c.guildId)
}

func (c *DashboardContext) Member() (member.Member, error) {
	return c.Worker().GetGuildMember(c.guildId, c.userId)
}

func (c *DashboardContext) User() (user.User, error) {
	return c.Worker().GetUser(c.UserId())
}

func (c *DashboardContext) IsBlacklisted(ctx context.Context) (bool, error) {
	permLevel, err := c.UserPermissionLevel(ctx)
	if err != nil {
		return false, err
	}

	member, err := c.Member()
	if err != nil {
		return false, err
	}

	// if interaction.Member is nil, it does not matter, as the member's roles are not checked
	// if the command is not executed in a guild
	return utils.IsBlacklisted(ctx, c.GuildId(), c.UserId(), member, permLevel)
}
