package command

import (
	"errors"
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
	"go.uber.org/atomic"
)

type SlashCommandContext struct {
	worker      *worker.Context
	Interaction interaction.ApplicationCommandInteraction
	premium     premium.PremiumTier

	hasReplied *atomic.Bool
	responseCh chan interaction.ApplicationCommandCallbackData
}

func NewSlashCommandContext(
	worker *worker.Context,
	interaction interaction.ApplicationCommandInteraction,
	premium premium.PremiumTier,
	responseCh chan interaction.ApplicationCommandCallbackData,
) SlashCommandContext {
	return SlashCommandContext{
		worker:      worker,
		Interaction: interaction,
		premium:     premium,

		hasReplied: atomic.NewBool(false),
		responseCh: responseCh,
	}
}

func (ctx *SlashCommandContext) Worker() *worker.Context {
	return ctx.worker
}

func (ctx *SlashCommandContext) GuildId() uint64 {
	return ctx.Interaction.GuildId.Value // TODO: Null check
}

func (ctx *SlashCommandContext) ChannelId() uint64 {
	return ctx.Interaction.ChannelId
}

func (ctx *SlashCommandContext) UserId() uint64 {
	return ctx.Interaction.Member.User.Id
}

func (ctx *SlashCommandContext) UserPermissionLevel() (permcache.PermissionLevel, error) {
	if ctx.Interaction.Member == nil {
		return permcache.Everyone, errors.New("member was nil")
	}

	return permcache.GetPermissionLevel(utils.ToRetriever(ctx.worker), *ctx.Interaction.Member, ctx.GuildId())
}

func (ctx *SlashCommandContext) PremiumTier() premium.PremiumTier {
	return ctx.premium
}

func (ctx *SlashCommandContext) IsInteraction() bool {
	return true
}

func (ctx *SlashCommandContext) ToErrorContext() errorcontext.WorkerErrorContext {
	return errorcontext.WorkerErrorContext{
		Guild:   ctx.GuildId(),
		User:    ctx.Interaction.Member.User.Id,
		Channel: ctx.ChannelId(),
	}
}

func (ctx *SlashCommandContext) reply(flags uint, content *embed.Embed) (message.Message, error) {
	hasReplied := ctx.hasReplied.Swap(true)

	if hasReplied {
		// TODO: Should we wait?
		msg, err := ctx.worker.ExecuteWebhook(ctx.worker.BotId, ctx.Interaction.Token, true, rest.WebhookBody{
			Embeds: []*embed.Embed{content},
			Flags:  flags,
		})

		if err != nil {
			sentry.LogWithContext(err, ctx.ToErrorContext())
		}

		if msg == nil {
			return message.Message{}, errors.New("message was nil")
		} else {
			return *msg, err
		}
	} else {
		ctx.responseCh <- interaction.ApplicationCommandCallbackData{
			Embeds: []*embed.Embed{content},
			Flags:  flags,
		}

		// todo: uhm
		return message.Message{}, nil
	}
}

func (ctx *SlashCommandContext) replyRaw(flags uint, content string) {
	hasReplied := ctx.hasReplied.Swap(true)

	if hasReplied {
		// TODO: Should we wait?
		_, err := ctx.worker.ExecuteWebhook(ctx.worker.BotId, ctx.Interaction.Token, false, rest.WebhookBody{
			Content: content,
			Flags:  flags,
		})

		if err != nil {
			sentry.LogWithContext(err, ctx.ToErrorContext())
		}
	} else {
		ctx.responseCh <- interaction.ApplicationCommandCallbackData{
			Content: content,
			Flags:  flags,
		}
	}
}

func (ctx *SlashCommandContext) buildEmbed(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) *embed.Embed {
	return utils.BuildEmbed(ctx.worker, ctx.GuildId(), colour, title, content, fields, ctx.premium > premium.None, format...)
}

func (ctx *SlashCommandContext) buildEmbedRaw(colour utils.Colour, title, content string, fields ...embed.EmbedField) *embed.Embed {
	return utils.BuildEmbedRaw(ctx.worker, colour, title, content, fields, ctx.premium > premium.None)
}

func (ctx *SlashCommandContext) Reply(colour utils.Colour, title string, content translations.MessageId, format ...interface{}) {
	embed := ctx.buildEmbed(colour, title, content, nil, format...)
	_, _  = ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *SlashCommandContext) ReplyWithEmbed(embed *embed.Embed) {
	_, _  = ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *SlashCommandContext) ReplyWithEmbedPermanent(embed *embed.Embed) (message.Message, error) {
	return ctx.reply(message.SumFlags(), embed)
}

func (ctx *SlashCommandContext) ReplyPermanent(colour utils.Colour, title string, content translations.MessageId, format ...interface{}) {
	embed := ctx.buildEmbed(colour, title, content, nil, format...)
	_, _  = ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *SlashCommandContext) ReplyWithFields(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) {
	embed := ctx.buildEmbed(colour, title, content, fields, format...)
	_, _  = ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *SlashCommandContext) ReplyWithFieldsPermanent(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) {
	embed := ctx.buildEmbed(colour, title, content, fields, format...)
	_, _  = ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *SlashCommandContext) ReplyRaw(colour utils.Colour, title, content string) {
	embed := ctx.buildEmbedRaw(colour, title, content)
	_, _  = ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *SlashCommandContext) ReplyRawPermanent(colour utils.Colour, title, content string) {
	embed := ctx.buildEmbedRaw(colour, title, content)
	_, _  = ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *SlashCommandContext) ReplyPlain(content string) {
	ctx.replyRaw(message.SumFlags(message.FlagEphemeral), content)
}

func (ctx *SlashCommandContext) ReplyPlainPermanent(content string) {
	ctx.replyRaw(0, content)
}

func (ctx *SlashCommandContext) Accept() {
	//ctx.ReplyPlain("✅")
}

func (ctx *SlashCommandContext) Reject() {
	//ctx.ReplyPlain("❌")
}

func (ctx *SlashCommandContext) HandleError(err error) {
	sentry.ErrorWithContext(err, ctx.ToErrorContext())

	embed := ctx.buildEmbedRaw(utils.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()))
	_, _  = ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *SlashCommandContext) HandleWarning(err error) {
	sentry.LogWithContext(err, ctx.ToErrorContext())

	embed := ctx.buildEmbedRaw(utils.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()))
	_, _  = ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *SlashCommandContext) Guild() (guild.Guild, error) {
	return ctx.Worker().GetGuild(ctx.GuildId())
}

func (ctx *SlashCommandContext) Member() (member.Member, error) {
	if ctx.Interaction.Member != nil {
		return *ctx.Interaction.Member, nil
	} else {
		return ctx.Worker().GetGuildMember(ctx.GuildId(), ctx.UserId())
	}
}

func (ctx *SlashCommandContext) User() (user.User, error) {
	return ctx.Worker().GetUser(ctx.UserId())
}
