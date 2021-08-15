package context

import (
	"fmt"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
)

type Replyable struct {
	ctx registry.CommandContext
}

func NewReplyable(ctx registry.CommandContext) *Replyable {
	return &Replyable{
		ctx: ctx,
	}
}

func (r *Replyable) buildEmbed(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) *embed.Embed {
	return utils.BuildEmbed(r.ctx.Worker(), r.ctx.GuildId(), colour, title, content, fields, r.ctx.PremiumTier() > premium.None, format...)
}

func (r *Replyable) buildEmbedRaw(colour utils.Colour, title, content string, fields ...embed.EmbedField) *embed.Embed {
	return utils.BuildEmbedRaw(r.ctx.Worker(), colour, title, content, fields, r.ctx.PremiumTier() > premium.None)
}

func (r *Replyable) Reply(colour utils.Colour, title string, content translations.MessageId, format ...interface{}) {
	embed := r.buildEmbed(colour, title, content, nil, format...)
	_, _ = r.ctx.ReplyWith(registry.NewEphemeralEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyPermanent(colour utils.Colour, title string, content translations.MessageId, format ...interface{}) {
	embed := r.buildEmbed(colour, title, content, nil, format...)
	_, _ = r.ctx.ReplyWith(registry.NewEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyWithEmbed(embed *embed.Embed) {
	_, _ = r.ctx.ReplyWith(registry.NewEphemeralEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyWithEmbedPermanent(embed *embed.Embed) {
	_, _ = r.ctx.ReplyWith(registry.NewEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyWithFields(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) {
	embed := r.buildEmbed(colour, title, content, fields, format...)
	_, _ = r.ctx.ReplyWith(registry.NewEphemeralEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyWithFieldsPermanent(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) {
	embed := r.buildEmbed(colour, title, content, fields, format...)
	_, _ = r.ctx.ReplyWith(registry.NewEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyRaw(colour utils.Colour, title, content string) {
	embed := r.buildEmbedRaw(colour, title, content)
	_, _ = r.ctx.ReplyWith(registry.NewEphemeralEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyRawPermanent(colour utils.Colour, title, content string) {
	embed := r.buildEmbedRaw(colour, title, content)
	_, _ = r.ctx.ReplyWith(registry.NewEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyPlain(content string) {
	_, _ = r.ctx.ReplyWith(registry.NewEphemeralTextMessageResponse(content))
}

func (r *Replyable) ReplyPlainPermanent(content string) {
	_, _ = r.ctx.ReplyWith(registry.NewTextMessageResponse(content))
}

func (r *Replyable) HandleError(err error) {
	sentry.ErrorWithContext(err, r.ctx.ToErrorContext())

	embed := r.buildEmbedRaw(utils.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()))
	_, _ = r.ctx.ReplyWith(registry.NewEphemeralEmbedMessageResponse(embed))
}

func (r *Replyable) HandleWarning(err error) {
	sentry.LogWithContext(err, r.ctx.ToErrorContext())

	embed := r.buildEmbedRaw(utils.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()))
	_, _ = r.ctx.ReplyWith(registry.NewEphemeralEmbedMessageResponse(embed))
}
