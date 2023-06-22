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

type PanelContext struct {
	*Replyable
	worker                     *worker.Context
	guildId, channelId, userId uint64
	premium                    premium.PremiumTier
	dmChannelId                uint64
}

func NewPanelContext(
	worker *worker.Context,
	guildId, channelId, userId uint64,
	premium premium.PremiumTier,
) PanelContext {
	ctx := PanelContext{
		worker:      worker,
		guildId:     guildId,
		channelId:   channelId,
		userId:      userId,
		premium:     premium,
		dmChannelId: 0,
	}

	ctx.Replyable = NewReplyable(&ctx)
	return ctx
}

func (ctx *PanelContext) Worker() *worker.Context {
	return ctx.worker
}

func (ctx *PanelContext) GuildId() uint64 {
	return ctx.guildId
}

func (ctx *PanelContext) ChannelId() uint64 {
	return ctx.channelId
}

func (ctx *PanelContext) UserId() uint64 {
	return ctx.userId
}

func (ctx *PanelContext) UserPermissionLevel() (permcache.PermissionLevel, error) {
	member, err := ctx.Member()
	if err != nil {
		return permcache.Everyone, err
	}

	return permcache.GetPermissionLevel(utils.ToRetriever(ctx.worker), member, ctx.guildId)
}

func (ctx *PanelContext) PremiumTier() premium.PremiumTier {
	return ctx.premium
}

func (ctx *PanelContext) IsInteraction() bool {
	return true
}

func (ctx *PanelContext) ToErrorContext() errorcontext.WorkerErrorContext {
	return errorcontext.WorkerErrorContext{
		Guild:   ctx.guildId,
		User:    ctx.userId,
		Channel: ctx.channelId,
	}
}

func (ctx *PanelContext) openDm() (uint64, bool) {
	if ctx.dmChannelId == 0 {
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

		ctx.dmChannelId = ch.Id
	}

	return ctx.dmChannelId, true
}

func (ctx *PanelContext) ReplyWith(response command.MessageResponse) (message.Message, error) {
	ch, ok := ctx.openDm()
	if !ok { // Error handled in openDm function
		return message.Message{}, errors.New("failed to open dm")
	}

	msg, err := ctx.Worker().CreateMessageComplex(ch, response.IntoCreateMessageData())
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	return msg, err
}

func (ctx *PanelContext) Accept() {}
func (ctx *PanelContext) Reject() {}

func (ctx *PanelContext) Channel() (channel.PartialChannel, error) {
	ch, err := ctx.Worker().GetChannel(ctx.channelId)
	if err != nil {
		return channel.PartialChannel{}, err
	}

	return ch.ToPartialChannel(), nil
}

func (ctx *PanelContext) Guild() (guild.Guild, error) {
	return ctx.Worker().GetGuild(ctx.guildId)
}

func (ctx *PanelContext) Member() (member.Member, error) {
	return ctx.Worker().GetGuildMember(ctx.guildId, ctx.userId)
}

func (ctx *PanelContext) User() (user.User, error) {
	return ctx.Worker().GetUser(ctx.UserId())
}

func (ctx *PanelContext) IsBlacklisted() (bool, error) {
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
