package command

import (
	"fmt"
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
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/rest"
)

type InteractionContext struct {
	worker             *worker.Context
	Interaction        interaction.Interaction
	premium            premium.PremiumTier
}

func NewInteractionContext(
	worker *worker.Context,
	interaction interaction.Interaction,
	premium premium.PremiumTier,
) InteractionContext {
	return InteractionContext{
		worker,
		interaction,
		premium,
	}
}

func (ctx *InteractionContext) Worker() *worker.Context {
	return ctx.worker
}

func (ctx *InteractionContext) GuildId() uint64 {
	return ctx.Interaction.GuildId
}

func (ctx *InteractionContext) ChannelId() uint64 {
	return ctx.Interaction.ChannelId
}

func (ctx *InteractionContext) UserId() uint64 {
	return ctx.Interaction.Member.User.Id
}

func (ctx *InteractionContext) UserPermissionLevel() permcache.PermissionLevel {
	return permcache.GetPermissionLevel(utils.ToRetriever(ctx.worker), ctx.Interaction.Member, ctx.GuildId())
}

func (ctx *InteractionContext) PremiumTier() premium.PremiumTier {
	return ctx.premium
}

func (ctx *InteractionContext) IsInteraction() bool {
	return true
}

func (ctx *InteractionContext) ToErrorContext() errorcontext.WorkerErrorContext {
	return errorcontext.WorkerErrorContext{
		Guild:   ctx.GuildId(),
		User:    ctx.Interaction.Member.User.Id,
		Channel: ctx.ChannelId(),
	}
}

func (ctx *InteractionContext) reply(flags uint, content *embed.Embed) {
	// TODO: Should we wait?
	_, err := ctx.worker.ExecuteWebhook(ctx.worker.BotId, ctx.Interaction.Token, false, rest.WebhookBody{
		Embeds: []*embed.Embed{content},
		Flags:  flags,
	})

	if err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext())
	}
}

func (ctx *InteractionContext) replyRaw(flags uint, content string) {
	// TODO: Should we wait?
	_, err := ctx.worker.ExecuteWebhook(ctx.worker.BotId, ctx.Interaction.Token, false, rest.WebhookBody{
		Content: content,
		Flags:   flags,
	})

	if err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext())
	}
}

func (ctx *InteractionContext) buildEmbed(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) *embed.Embed {
	return utils.BuildEmbed(ctx.worker, ctx.GuildId(), colour, title, content, fields, ctx.premium > premium.None, format...)
}

func (ctx *InteractionContext) buildEmbedRaw(colour utils.Colour, title, content string, fields ...embed.EmbedField) *embed.Embed {
	return utils.BuildEmbedRaw(ctx.worker, colour, title, content, fields, ctx.premium > premium.None)
}

func (ctx *InteractionContext) Reply(colour utils.Colour, title string, content translations.MessageId, format ...interface{}) {
	embed := ctx.buildEmbed(colour, title, content, nil, format...)
	ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *InteractionContext) ReplyWithEmbed(embed *embed.Embed) {
	ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *InteractionContext) ReplyPermanent(colour utils.Colour, title string, content translations.MessageId, format ...interface{}) {
	embed := ctx.buildEmbed(colour, title, content, nil, format...)
	ctx.reply(message.SumFlags(), embed)
}

func (ctx *InteractionContext) ReplyWithFields(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) {
	embed := ctx.buildEmbed(colour, title, content, fields, format...)
	ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *InteractionContext) ReplyRaw(colour utils.Colour, title, content string) {
	embed := ctx.buildEmbedRaw(colour, title, content)
	ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *InteractionContext) ReplyRawPermanent(colour utils.Colour, title, content string) {
	embed := ctx.buildEmbedRaw(colour, title, content)
	ctx.reply(message.SumFlags(), embed)
}

func (ctx *InteractionContext) ReplyPlain(content string) {
	ctx.replyRaw(message.SumFlags(message.FlagEphemeral), content)
}

func (ctx *InteractionContext) Accept() {}
func (ctx *InteractionContext) Reject() {}

func (ctx *InteractionContext) HandleError(err error) {
	sentry.ErrorWithContext(err, ctx.ToErrorContext())

	embed := ctx.buildEmbedRaw(utils.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()))
	ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *InteractionContext) HandleWarning(err error) {
	sentry.LogWithContext(err, ctx.ToErrorContext())

	embed := ctx.buildEmbedRaw(utils.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()))
	ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *InteractionContext) Guild() (guild.Guild, error) {
	return ctx.Worker().GetGuild(ctx.GuildId())
}

func (ctx *InteractionContext) Member() (member.Member, error) {
	return ctx.Interaction.Member, nil
}

func (ctx *InteractionContext) User() (user.User, error) {
	return ctx.Worker().GetUser(ctx.UserId())
}
