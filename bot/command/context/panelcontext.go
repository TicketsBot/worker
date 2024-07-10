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

type PanelContext struct {
	context.Context
	*Replyable
	*StateCache
	worker                     *worker.Context
	guildId, channelId, userId uint64
	premium                    premium.PremiumTier
	dmChannelId                uint64
}

var _ registry.CommandContext = (*PanelContext)(nil)

func NewPanelContext(
	ctx context.Context,
	worker *worker.Context,
	guildId, channelId, userId uint64,
	premium premium.PremiumTier,
) PanelContext {
	c := PanelContext{
		Context:     ctx,
		worker:      worker,
		guildId:     guildId,
		channelId:   channelId,
		userId:      userId,
		premium:     premium,
		dmChannelId: 0,
	}

	c.Replyable = NewReplyable(&c)
	c.StateCache = NewStateCache(&c)
	return c
}

func (c *PanelContext) Worker() *worker.Context {
	return c.worker
}

func (c *PanelContext) GuildId() uint64 {
	return c.guildId
}

func (c *PanelContext) ChannelId() uint64 {
	return c.channelId
}

func (c *PanelContext) UserId() uint64 {
	return c.userId
}

func (c *PanelContext) UserPermissionLevel(ctx context.Context) (permcache.PermissionLevel, error) {
	member, err := c.Member()
	if err != nil {
		return permcache.Everyone, err
	}

	return permcache.GetPermissionLevel(ctx, utils.ToRetriever(c.worker), member, c.guildId)
}

func (c *PanelContext) PremiumTier() premium.PremiumTier {
	return c.premium
}

func (c *PanelContext) IsInteraction() bool {
	return true
}

func (c *PanelContext) Source() registry.Source {
	return registry.SourceDashboard // TODO: Correct source?
}

func (c *PanelContext) ToErrorContext() errorcontext.WorkerErrorContext {
	return errorcontext.WorkerErrorContext{
		Guild:   c.guildId,
		User:    c.userId,
		Channel: c.channelId,
	}
}

func (c *PanelContext) openDm() (uint64, bool) {
	if c.dmChannelId == 0 {
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

		c.dmChannelId = ch.Id
	}

	return c.dmChannelId, true
}

func (c *PanelContext) ReplyWith(response command.MessageResponse) (message.Message, error) {
	ch, ok := c.openDm()
	if !ok { // Error handled in openDm function
		return message.Message{}, errors.New("failed to open dm")
	}

	msg, err := c.Worker().CreateMessageComplex(ch, response.IntoCreateMessageData())
	if err != nil {
		sentry.ErrorWithContext(err, c.ToErrorContext())
	}

	return msg, err
}

func (c *PanelContext) Channel() (channel.PartialChannel, error) {
	ch, err := c.Worker().GetChannel(c.channelId)
	if err != nil {
		return channel.PartialChannel{}, err
	}

	return ch.ToPartialChannel(), nil
}

func (c *PanelContext) Guild() (guild.Guild, error) {
	return c.Worker().GetGuild(c.guildId)
}

func (c *PanelContext) Member() (member.Member, error) {
	return c.Worker().GetGuildMember(c.guildId, c.userId)
}

func (c *PanelContext) User() (user.User, error) {
	return c.Worker().GetUser(c.UserId())
}

func (c *PanelContext) IsBlacklisted(ctx context.Context) (bool, error) {
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
