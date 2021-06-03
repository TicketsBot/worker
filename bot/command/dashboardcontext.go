package command

import (
	"fmt"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/rest/request"
)

type DashboardContext struct {
	worker                     *worker.Context
	guildId, channelId, userId uint64
	premium                    premium.PremiumTier
}

func NewDashboardContext(
	worker *worker.Context,
	guildId, channelId, userId uint64,
	premium premium.PremiumTier,
) DashboardContext {
	return DashboardContext{
		worker,
		guildId, channelId, userId,
		premium,
	}
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
	cachedId, err := redis.GetDMChannel(ctx.UserId())
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
			if err := redis.StoreNullDMChannel(ctx.UserId()); err != nil {
				sentry.ErrorWithContext(err, ctx.ToErrorContext())
			}

			return 0, false
		}

		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return 0, false
	}

	if err := redis.StoreDMChannel(ctx.UserId(), ch.Id); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	return ch.Id, true
}

func (ctx *DashboardContext) reply(content *embed.Embed) {
	channelId, ok := ctx.openDm()
	if !ok { // Error handled in openDm function
		return
	}

	if _, err := ctx.Worker().CreateMessageEmbed(channelId, content); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}
}

func (ctx *DashboardContext) replyRaw(content string) {
	channelId, ok := ctx.openDm()
	if !ok { // Error handled in openDm function
		return
	}

	if _, err := ctx.Worker().CreateMessage(channelId, content); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}
}

func (ctx *DashboardContext) buildEmbed(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) *embed.Embed {
	return utils.BuildEmbed(ctx.worker, ctx.guildId, colour, title, content, fields, ctx.premium > premium.None, format...)
}

func (ctx *DashboardContext) buildEmbedRaw(colour utils.Colour, title, content string, fields ...embed.EmbedField) *embed.Embed {
	return utils.BuildEmbedRaw(ctx.worker, colour, title, content, fields, ctx.premium > premium.None)
}

func (ctx *DashboardContext) Reply(colour utils.Colour, title string, content translations.MessageId, format ...interface{}) {
	embed := ctx.buildEmbed(colour, title, content, nil, format...)
	ctx.reply(embed)
}

func (ctx *DashboardContext) ReplyWithEmbed(embed *embed.Embed) {
	ctx.reply(embed)
}

func (ctx *DashboardContext) ReplyWithEmbedPermanent(embed *embed.Embed) {
	ctx.reply(embed)
}

func (ctx *DashboardContext) ReplyPermanent(colour utils.Colour, title string, content translations.MessageId, format ...interface{}) {
	embed := ctx.buildEmbed(colour, title, content, nil, format...)
	ctx.reply(embed)
}

func (ctx *DashboardContext) ReplyWithFields(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) {
	embed := ctx.buildEmbed(colour, title, content, fields, format...)
	ctx.reply(embed)
}

func (ctx *DashboardContext) ReplyWithFieldsPermanent(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) {
	embed := ctx.buildEmbed(colour, title, content, fields, format...)
	ctx.reply(embed)
}

func (ctx *DashboardContext) ReplyRaw(colour utils.Colour, title, content string) {
	embed := ctx.buildEmbedRaw(colour, title, content)
	ctx.reply(embed)
}

func (ctx *DashboardContext) ReplyRawPermanent(colour utils.Colour, title, content string) {
	embed := ctx.buildEmbedRaw(colour, title, content)
	ctx.reply(embed)}

func (ctx *DashboardContext) ReplyPlain(content string) {
	ctx.replyRaw(content)
}

func (ctx *DashboardContext) ReplyPlainPermanent(content string) {
	ctx.replyRaw(content)
}

func (ctx *DashboardContext) Accept() {}
func (ctx *DashboardContext) Reject() {}

func (ctx *DashboardContext) HandleError(err error) {
	sentry.ErrorWithContext(err, ctx.ToErrorContext())

	embed := ctx.buildEmbedRaw(utils.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()))
	ctx.reply(embed)
}

func (ctx *DashboardContext) HandleWarning(err error) {
	sentry.LogWithContext(err, ctx.ToErrorContext())

	embed := ctx.buildEmbedRaw(utils.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()))
	ctx.reply(embed)
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
