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
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
)

type PanelContext struct {
	worker                     *worker.Context
	guildId, channelId, userId uint64
	premium                    premium.PremiumTier
}

func NewPanelContext(
	worker *worker.Context,
	guildId, channelId, userId uint64,
	premium premium.PremiumTier,
) PanelContext {
	return PanelContext{
		worker,
		guildId, channelId, userId,
		premium,
	}
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

func (ctx *PanelContext) UserPermissionLevel() permcache.PermissionLevel {
	member, err := ctx.Member()
	if err != nil {
		ctx.HandleError(err)
		return permcache.Everyone
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

// TODO: DM user
func (ctx *PanelContext) reply(content *embed.Embed) {}
func (ctx *PanelContext) replyRaw(content string)    {}

func (ctx *PanelContext) buildEmbed(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) *embed.Embed {
	return utils.BuildEmbed(ctx.worker, ctx.guildId, colour, title, content, fields, ctx.premium > premium.None, format...)
}

func (ctx *PanelContext) buildEmbedRaw(colour utils.Colour, title, content string, fields ...embed.EmbedField) *embed.Embed {
	return utils.BuildEmbedRaw(ctx.worker, colour, title, content, fields, ctx.premium > premium.None)
}

func (ctx *PanelContext) Reply(colour utils.Colour, title string, content translations.MessageId, format ...interface{}) {
}

func (ctx *PanelContext) ReplyWithEmbed(embed *embed.Embed) {}

func (ctx *PanelContext) ReplyPermanent(colour utils.Colour, title string, content translations.MessageId, format ...interface{}) {
}

func (ctx *PanelContext) ReplyWithFields(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) {
}

func (ctx *PanelContext) ReplyRaw(colour utils.Colour, title, content string) {}

func (ctx *PanelContext) ReplyRawPermanent(colour utils.Colour, title, content string) {}

func (ctx *PanelContext) ReplyPlain(content string) {}

func (ctx *PanelContext) Accept() {}
func (ctx *PanelContext) Reject() {}

func (ctx *PanelContext) HandleError(err error) {
	sentry.ErrorWithContext(err, ctx.ToErrorContext())
}

func (ctx *PanelContext) HandleWarning(err error) {
	sentry.LogWithContext(err, ctx.ToErrorContext())
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
