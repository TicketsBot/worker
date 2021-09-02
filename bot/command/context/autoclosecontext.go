package context

import (
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
)

type AutoCloseContext struct {
	*Replyable
	worker                     *worker.Context
	guildId, channelId, userId uint64
	premium                    premium.PremiumTier
}

func NewAutoCloseContext(
	worker *worker.Context,
	guildId, channelId, userId uint64,
	premium premium.PremiumTier,
) AutoCloseContext {
	ctx := AutoCloseContext{
		worker:    worker,
		guildId:   guildId,
		channelId: channelId,
		userId:    userId,
		premium:   premium,
	}

	ctx.Replyable = NewReplyable(&ctx)
	return ctx
}

func (ctx *AutoCloseContext) Worker() *worker.Context {
	return ctx.worker
}

func (ctx *AutoCloseContext) GuildId() uint64 {
	return ctx.guildId
}

func (ctx *AutoCloseContext) ChannelId() uint64 {
	return ctx.channelId
}

func (ctx *AutoCloseContext) UserId() uint64 {
	return ctx.userId
}

// TODO: Could this be dangerous? Don't think so, since this context is only used for closing
func (ctx *AutoCloseContext) UserPermissionLevel() (permcache.PermissionLevel, error) {
	return permcache.Admin, nil
}

func (ctx *AutoCloseContext) PremiumTier() premium.PremiumTier {
	return ctx.premium
}

func (ctx *AutoCloseContext) IsInteraction() bool {
	return true
}

func (ctx *AutoCloseContext) ToErrorContext() errorcontext.WorkerErrorContext {
	return errorcontext.WorkerErrorContext{
		Guild:   ctx.guildId,
		User:    ctx.userId,
		Channel: ctx.channelId,
	}
}

func (ctx *AutoCloseContext) openDm() (uint64, bool) {
	return 0, false
}

func (ctx *AutoCloseContext) ReplyWith(response command.MessageResponse) (message.Message, error) {
	return message.Message{}, nil
}

func (ctx *AutoCloseContext) Accept() {}
func (ctx *AutoCloseContext) Reject() {}

func (ctx *AutoCloseContext) Guild() (guild.Guild, error) {
	return ctx.Worker().GetGuild(ctx.guildId)
}

func (ctx *AutoCloseContext) Member() (member.Member, error) {
	return ctx.Worker().GetGuildMember(ctx.guildId, ctx.userId)
}

func (ctx *AutoCloseContext) User() (user.User, error) {
	return ctx.Worker().GetUser(ctx.UserId())
}
