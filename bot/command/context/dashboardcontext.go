package context

import (
	"errors"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
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
	*Replyable
	*StateCache
	worker                     *worker.Context
	guildId, channelId, userId uint64
	premium                    premium.PremiumTier
}

func NewDashboardContext(
	worker *worker.Context,
	guildId, channelId, userId uint64,
	premium premium.PremiumTier,
) DashboardContext {
	ctx := DashboardContext{
		worker:    worker,
		guildId:   guildId,
		channelId: channelId,
		userId:    userId,
		premium:   premium,
	}

	ctx.Replyable = NewReplyable(&ctx)
	ctx.StateCache = NewStateCache(&ctx)
	return ctx
}

func (ctx *DashboardContext) Worker() *worker.Context {
	return ctx.worker
}

func (ctx *DashboardContext) GuildId() uint64 {
	return ctx.guildId
}

func (ctx *DashboardContext) ChannelId() uint64 {
	return ctx.channelId
}

func (ctx *DashboardContext) UserId() uint64 {
	return ctx.userId
}

func (ctx *DashboardContext) UserPermissionLevel() (permcache.PermissionLevel, error) {
	member, err := ctx.Member()
	if err != nil {
		return permcache.Everyone, err
	}

	return permcache.GetPermissionLevel(utils.ToRetriever(ctx.worker), member, ctx.guildId)
}

func (ctx *DashboardContext) PremiumTier() premium.PremiumTier {
	return ctx.premium
}

func (ctx *DashboardContext) IsInteraction() bool {
	return true
}

func (ctx *DashboardContext) ToErrorContext() errorcontext.WorkerErrorContext {
	return errorcontext.WorkerErrorContext{
		Guild:   ctx.guildId,
		User:    ctx.userId,
		Channel: ctx.channelId,
	}
}

func (ctx *DashboardContext) openDm() (uint64, bool) {
	cachedId, err := redis.GetDMChannel(ctx.UserId(), ctx.Worker().BotId)
	if err != nil { // We can continue
		if err != redis.ErrNotCached {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		}
	} else { // We have it cached
		if cachedId == nil {
			return 0, false
		} else {
			return *cachedId, true
		}
	}

	ch, err := ctx.Worker().CreateDM(ctx.UserId())
	if err != nil {
		// check for 403
		if err, ok := err.(request.RestError); ok && err.StatusCode == 403 {
			if err := redis.StoreNullDMChannel(ctx.UserId(), ctx.Worker().BotId); err != nil {
				sentry.ErrorWithContext(err, ctx.ToErrorContext())
			}

			return 0, false
		}

		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return 0, false
	}

	if err := redis.StoreDMChannel(ctx.UserId(), ch.Id, ctx.Worker().BotId); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	return ch.Id, true
}

func (ctx *DashboardContext) ReplyWith(response command.MessageResponse) (message.Message, error) {
	channelId, ok := ctx.openDm()
	if !ok { // Error handled in openDm function
		return message.Message{}, errors.New("failed to open dm")
	}

	msg, err := ctx.Worker().CreateMessageComplex(channelId, response.IntoCreateMessageData())
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	return msg, err
}

func (ctx *DashboardContext) Channel() (channel.PartialChannel, error) {
	ch, err := ctx.Worker().GetChannel(ctx.channelId)
	if err != nil {
		return channel.PartialChannel{}, err
	}

	return ch.ToPartialChannel(), nil
}

func (ctx *DashboardContext) Guild() (guild.Guild, error) {
	return ctx.Worker().GetGuild(ctx.guildId)
}

func (ctx *DashboardContext) Member() (member.Member, error) {
	return ctx.Worker().GetGuildMember(ctx.guildId, ctx.userId)
}

func (ctx *DashboardContext) User() (user.User, error) {
	return ctx.Worker().GetUser(ctx.UserId())
}

func (ctx *DashboardContext) IsBlacklisted() (bool, error) {
	permLevel, err := ctx.UserPermissionLevel()
	if err != nil {
		return false, err
	}

	member, err := ctx.Member()
	if err != nil {
		return false, err
	}

	// if interaction.Member is nil, it does not matter, as the member's roles are not checked
	// if the command is not executed in a guild
	return utils.IsBlacklisted(ctx.GuildId(), ctx.UserId(), member, permLevel)
}
