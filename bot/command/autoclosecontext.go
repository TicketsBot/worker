package command

import (
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
)

type AutoCloseContext struct {
	worker                     *worker.Context
	guildId, channelId, userId uint64
	premium                    premium.PremiumTier
}

func NewAutoCloseContext(
	worker *worker.Context,
	guildId, channelId, userId uint64,
	premium premium.PremiumTier,
) AutoCloseContext {
	return AutoCloseContext{
		worker,
		guildId, channelId, userId,
		premium,
	}
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

func (ctx *AutoCloseContext) reply(content *embed.Embed) (message.Message, error) {
	return message.Message{}, nil
}

func (ctx *AutoCloseContext) replyRaw(content string) {
	return
}

func (ctx *AutoCloseContext) buildEmbed(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) *embed.Embed {
	return utils.BuildEmbed(ctx.worker, ctx.guildId, colour, title, content, fields, ctx.premium > premium.None, format...)
}

func (ctx *AutoCloseContext) buildEmbedRaw(colour utils.Colour, title, content string, fields ...embed.EmbedField) *embed.Embed {
	return utils.BuildEmbedRaw(ctx.worker, colour, title, content, fields, ctx.premium > premium.None)
}

func (ctx *AutoCloseContext) Reply(colour utils.Colour, title string, content translations.MessageId, format ...interface{}) {}

func (ctx *AutoCloseContext) ReplyWithEmbed(embed *embed.Embed) {}

func (ctx *AutoCloseContext) ReplyWithEmbedPermanent(embed *embed.Embed) (message.Message, error) {
	return ctx.reply(embed)
}

func (ctx *AutoCloseContext) ReplyPermanent(colour utils.Colour, title string, content translations.MessageId, format ...interface{}) {}

func (ctx *AutoCloseContext) ReplyWithFields(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) {}

func (ctx *AutoCloseContext) ReplyWithFieldsPermanent(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) {}

func (ctx *AutoCloseContext) ReplyRaw(colour utils.Colour, title, content string) {}

func (ctx *AutoCloseContext) ReplyRawPermanent(colour utils.Colour, title, content string) {}

func (ctx *AutoCloseContext) ReplyPlain(content string) {}

func (ctx *AutoCloseContext) ReplyPlainPermanent(content string) {}

func (ctx *AutoCloseContext) Accept() {}
func (ctx *AutoCloseContext) Reject() {}

func (ctx *AutoCloseContext) HandleError(err error) {
	sentry.ErrorWithContext(err, ctx.ToErrorContext())
}

func (ctx *AutoCloseContext) HandleWarning(err error) {
	sentry.LogWithContext(err, ctx.ToErrorContext())
}

func (ctx *AutoCloseContext) Guild() (guild.Guild, error) {
	return ctx.Worker().GetGuild(ctx.guildId)
}

func (ctx *AutoCloseContext) Member() (member.Member, error) {
	return ctx.Worker().GetGuildMember(ctx.guildId, ctx.userId)
}

func (ctx *AutoCloseContext) User() (user.User, error) {
	return ctx.Worker().GetUser(ctx.UserId())
}
