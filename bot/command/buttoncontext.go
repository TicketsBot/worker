package command

/*import (
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
)

type ButtonContext struct {
	worker      *worker.Context
	Interaction interaction.ButtonInteraction
	premium     premium.PremiumTier
}

func NewButtonContext(
	worker *worker.Context,
	interaction interaction.ButtonInteraction,
	premium premium.PremiumTier,
) ButtonContext {
	return ButtonContext{
		worker:      worker,
		Interaction: interaction,
		premium:     premium,
	}
}

func (ctx *ButtonContext) Worker() *worker.Context {
	return ctx.worker
}

func (ctx *ButtonContext) GuildId() uint64 {
	return ctx.Interaction.GuildId.Value // TODO: Null check
}

func (ctx *ButtonContext) ChannelId() uint64 {
	return ctx.Interaction.ChannelId
}

func (ctx *ButtonContext) UserId() uint64 {
	return ctx.Interaction.Member.User.Id
}

func (ctx *ButtonContext) UserPermissionLevel() (permcache.PermissionLevel, error) {
	if ctx.Interaction.Member == nil {
		return permcache.Everyone, errors.New("member was nil")
	}

	return permcache.GetPermissionLevel(utils.ToRetriever(ctx.worker), *ctx.Interaction.Member, ctx.GuildId())
}

func (ctx *ButtonContext) PremiumTier() premium.PremiumTier {
	return ctx.premium
}

func (ctx *ButtonContext) IsInteraction() bool {
	return true
}

func (ctx *ButtonContext) ToErrorContext() errorcontext.WorkerErrorContext {
	return errorcontext.WorkerErrorContext{
		Guild:   ctx.GuildId(),
		User:    ctx.Interaction.Member.User.Id,
		Channel: ctx.ChannelId(),
	}
}

func (ctx *ButtonContext) reply(flags uint, content *embed.Embed) (message.Message, error) {
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

func (ctx *ButtonContext) replyRaw(flags uint, content string) {
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

func (ctx *ButtonContext) buildEmbed(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) *embed.Embed {
	return utils.BuildEmbed(ctx.worker, ctx.GuildId(), colour, title, content, fields, ctx.premium > premium.None, format...)
}

func (ctx *ButtonContext) buildEmbedRaw(colour utils.Colour, title, content string, fields ...embed.EmbedField) *embed.Embed {
	return utils.BuildEmbedRaw(ctx.worker, colour, title, content, fields, ctx.premium > premium.None)
}

func (ctx *ButtonContext) Reply(colour utils.Colour, title string, content translations.MessageId, format ...interface{}) {
	embed := ctx.buildEmbed(colour, title, content, nil, format...)
	_, _  = ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *ButtonContext) ReplyWithEmbed(embed *embed.Embed) {
	_, _  = ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *ButtonContext) ReplyWithEmbedPermanent(embed *embed.Embed) (message.Message, error) {
	return ctx.reply(message.SumFlags(), embed)
}

func (ctx *ButtonContext) ReplyPermanent(colour utils.Colour, title string, content translations.MessageId, format ...interface{}) {
	embed := ctx.buildEmbed(colour, title, content, nil, format...)
	_, _  = ctx.reply(message.SumFlags(), embed)
}

func (ctx *ButtonContext) ReplyWithFields(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) {
	embed := ctx.buildEmbed(colour, title, content, fields, format...)
	_, _  = ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *ButtonContext) ReplyWithFieldsPermanent(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) {
	embed := ctx.buildEmbed(colour, title, content, fields, format...)
	_, _  = ctx.reply(message.SumFlags(), embed)
}

func (ctx *ButtonContext) ReplyRaw(colour utils.Colour, title, content string) {
	embed := ctx.buildEmbedRaw(colour, title, content)
	_, _  = ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *ButtonContext) ReplyRawPermanent(colour utils.Colour, title, content string) {
	embed := ctx.buildEmbedRaw(colour, title, content)
	_, _  = ctx.reply(message.SumFlags(), embed)
}

func (ctx *ButtonContext) ReplyPlain(content string) {
	ctx.replyRaw(message.SumFlags(message.FlagEphemeral), content)
}

func (ctx *ButtonContext) ReplyPlainPermanent(content string) {
	ctx.replyRaw(0, content)
}

func (ctx *ButtonContext) Accept() {
	//ctx.ReplyPlain("✅")
}

func (ctx *ButtonContext) Reject() {
	//ctx.ReplyPlain("❌")
}

func (ctx *ButtonContext) HandleError(err error) {
	sentry.ErrorWithContext(err, ctx.ToErrorContext())

	embed := ctx.buildEmbedRaw(utils.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()))
	_, _  = ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *ButtonContext) HandleWarning(err error) {
	sentry.LogWithContext(err, ctx.ToErrorContext())

	embed := ctx.buildEmbedRaw(utils.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()))
	_, _  = ctx.reply(message.SumFlags(message.FlagEphemeral), embed)
}

func (ctx *ButtonContext) Guild() (guild.Guild, error) {
	return ctx.Worker().GetGuild(ctx.GuildId())
}

func (ctx *ButtonContext) Member() (member.Member, error) {
	if ctx.Interaction.Member != nil {
		return *ctx.Interaction.Member, nil
	} else {
		return ctx.Worker().GetGuildMember(ctx.GuildId(), ctx.UserId())
	}
}

func (ctx *ButtonContext) User() (user.User, error) {
	return ctx.Worker().GetUser(ctx.UserId())
}
*/