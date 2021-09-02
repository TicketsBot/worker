package context

import (
	"fmt"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
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

func (r *Replyable) buildEmbed(colour constants.Colour, title string, content i18n.MessageId, fields []embed.EmbedField, format ...interface{}) *embed.Embed {
	return utils.BuildEmbed(r.ctx, colour, title, content, fields, format...)
}

func (r *Replyable) buildEmbedRaw(colour constants.Colour, title, content string, fields ...embed.EmbedField) *embed.Embed {
	return utils.BuildEmbedRaw(colour, title, content, fields, r.ctx.PremiumTier())
}

func (r *Replyable) Reply(colour constants.Colour, title string, content i18n.MessageId, format ...interface{}) {
	embed := r.buildEmbed(colour, title, content, nil, format...)
	_, _ = r.ctx.ReplyWith(command.NewEphemeralEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyPermanent(colour constants.Colour, title string, content i18n.MessageId, format ...interface{}) {
	embed := r.buildEmbed(colour, title, content, nil, format...)
	_, _ = r.ctx.ReplyWith(command.NewEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyWithEmbed(embed *embed.Embed) {
	_, _ = r.ctx.ReplyWith(command.NewEphemeralEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyWithEmbedPermanent(embed *embed.Embed) {
	_, _ = r.ctx.ReplyWith(command.NewEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyWithFields(colour constants.Colour, title string, content i18n.MessageId, fields []embed.EmbedField, format ...interface{}) {
	embed := r.buildEmbed(colour, title, content, fields, format...)
	_, _ = r.ctx.ReplyWith(command.NewEphemeralEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyWithFieldsPermanent(colour constants.Colour, title string, content i18n.MessageId, fields []embed.EmbedField, format ...interface{}) {
	embed := r.buildEmbed(colour, title, content, fields, format...)
	_, _ = r.ctx.ReplyWith(command.NewEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyRaw(colour constants.Colour, title, content string) {
	embed := r.buildEmbedRaw(colour, title, content)
	_, _ = r.ctx.ReplyWith(command.NewEphemeralEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyRawPermanent(colour constants.Colour, title, content string) {
	embed := r.buildEmbedRaw(colour, title, content)
	_, _ = r.ctx.ReplyWith(command.NewEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyPlain(content string) {
	_, _ = r.ctx.ReplyWith(command.NewEphemeralTextMessageResponse(content))
}

func (r *Replyable) ReplyPlainPermanent(content string) {
	_, _ = r.ctx.ReplyWith(command.NewTextMessageResponse(content))
}

func (r *Replyable) HandleError(err error) {
	sentry.ErrorWithContext(err, r.ctx.ToErrorContext())

	embed := r.buildEmbedRaw(constants.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()))
	_, _ = r.ctx.ReplyWith(command.NewEphemeralEmbedMessageResponse(embed))
}

func (r *Replyable) HandleWarning(err error) {
	sentry.LogWithContext(err, r.ctx.ToErrorContext())

	embed := r.buildEmbedRaw(constants.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()))
	_, _ = r.ctx.ReplyWith(command.NewEphemeralEmbedMessageResponse(embed))
}
