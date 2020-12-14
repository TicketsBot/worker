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
)

type MessageContext struct {
	worker *worker.Context
	message.Message
	Args            []string
	premium         premium.PremiumTier
	permissionLevel permcache.PermissionLevel
}

func (ctx *MessageContext) Worker() *worker.Context {
	return ctx.worker
}

func (ctx *MessageContext) GuildId() uint64 {
	return ctx.Message.GuildId
}

func (ctx *MessageContext) ChannelId() uint64 {
	return ctx.Message.ChannelId
}

func (ctx *MessageContext) UserId() uint64 {
	return ctx.Author.Id
}

func (ctx *MessageContext) UserPermissionLevel() permcache.PermissionLevel {
	return ctx.permissionLevel
}

func (ctx *MessageContext) PremiumTier() premium.PremiumTier {
	return ctx.premium
}

func (ctx *MessageContext) IsInteraction() bool {
	return false
}

func (ctx *MessageContext) ToErrorContext() errorcontext.WorkerErrorContext {
	return errorcontext.WorkerErrorContext{
		Guild:   ctx.GuildId(),
		User:    ctx.Author.Id,
		Channel: ctx.ChannelId(),
		Shard:   ctx.worker.ShardId,
	}
}

func (ctx *MessageContext) ReplyContext() *message.MessageReference {
	return &message.MessageReference{
		MessageId: ctx.Id,
		ChannelId: ctx.ChannelId(),
		GuildId:   ctx.GuildId(),
	}
}

func (ctx *MessageContext) reply(content *embed.Embed) (message.Message, bool) {
	msg, err := ctx.worker.CreateMessageEmbedReply(ctx.ChannelId(), content, ctx.ReplyContext())

	if err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext())
	}

	return msg, err == nil
}

func (ctx *MessageContext) replyRaw(content string) (message.Message, bool) {
	msg, err := ctx.worker.CreateMessageReply(ctx.ChannelId(), content, ctx.ReplyContext())

	if err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext())
	}

	return msg, err == nil
}

func (ctx *MessageContext) buildEmbed(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) *embed.Embed {
	return utils.BuildEmbed(ctx.worker, ctx.GuildId(), colour, title, content, fields, ctx.premium > premium.None, format...)
}

func (ctx *MessageContext) buildEmbedRaw(colour utils.Colour, title, content string, fields ...embed.EmbedField) *embed.Embed {
	return utils.BuildEmbedRaw(ctx.worker, colour, title, content, fields, ctx.premium > premium.None)
}

func (ctx *MessageContext) Reply(colour utils.Colour, title string, content translations.MessageId, format ...interface{}) {
	if msg, ok := ctx.reply(ctx.buildEmbed(colour, title, content, nil, format...)); ok {
		utils.DeleteAfter(ctx.worker, msg.ChannelId, msg.Id, utils.DeleteAfterSeconds)
	}
}

func (ctx *MessageContext) ReplyWithEmbed(embed *embed.Embed) {
	ctx.reply(embed)
}

func (ctx *MessageContext) ReplyPermanent(colour utils.Colour, title string, content translations.MessageId, format ...interface{}) {
	ctx.reply(ctx.buildEmbed(colour, title, content, nil, format...))
}

func (ctx *MessageContext) ReplyWithFields(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) {
	if msg, ok := ctx.reply(ctx.buildEmbed(colour, title, content, fields, format...)); ok {
		utils.DeleteAfter(ctx.worker, msg.ChannelId, msg.Id, utils.DeleteAfterSeconds)
	}
}

func (ctx *MessageContext) ReplyRaw(colour utils.Colour, title, content string) {
	if msg, ok := ctx.reply(ctx.buildEmbedRaw(colour, title, content)); ok {
		utils.DeleteAfter(ctx.worker, msg.ChannelId, msg.Id, utils.DeleteAfterSeconds)
	}
}

func (ctx *MessageContext) ReplyRawPermanent(colour utils.Colour, title, content string) {
	ctx.reply(ctx.buildEmbedRaw(colour, title, content))
}

func (ctx *MessageContext) ReplyPlain(content string) {
	if msg, ok := ctx.replyRaw(content); ok {
		utils.DeleteAfter(ctx.worker, msg.ChannelId, msg.Id, utils.DeleteAfterSeconds)
	}
}

func (ctx *MessageContext) Accept() {
	utils.ReactWithCheck(ctx.worker, ctx.ChannelId(), ctx.Id)
}

func (ctx *MessageContext) Reject() {
	utils.ReactWithCross(ctx.worker, ctx.ChannelId(), ctx.Id)
}

func (ctx *MessageContext) HandleError(err error) {
	sentry.ErrorWithContext(err, ctx.ToErrorContext())

	embed := ctx.buildEmbedRaw(utils.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()))
	ctx.reply(embed)
}

func (ctx *MessageContext) HandleWarning(err error) {
	sentry.LogWithContext(err, ctx.ToErrorContext())

	embed := ctx.buildEmbedRaw(utils.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()))
	ctx.reply(embed)
}
