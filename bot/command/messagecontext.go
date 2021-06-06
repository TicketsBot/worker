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
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
)

type MessageContext struct {
	worker *worker.Context
	message.Message
	Args            []string
	premium         premium.PremiumTier
	permissionLevel permcache.PermissionLevel
}

func NewMessageContext(
	worker *worker.Context,
	message message.Message,
	args []string,
	premium premium.PremiumTier,
	permissionLevel permcache.PermissionLevel,
) MessageContext {
	return MessageContext{
		worker,
		message,
		args,
		premium,
		permissionLevel,
	}
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

func (ctx *MessageContext) UserPermissionLevel() (permcache.PermissionLevel, error) {
	return ctx.permissionLevel, nil
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

func (ctx *MessageContext) reply(content *embed.Embed) (message.Message, error) {
	msg, err := ctx.worker.CreateMessageEmbedReply(ctx.ChannelId(), content, ctx.ReplyContext())

	if err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext())
	}

	return msg, err
}

func (ctx *MessageContext) replyRaw(content string) (message.Message, error) {
	msg, err := ctx.worker.CreateMessageReply(ctx.ChannelId(), content, ctx.ReplyContext())

	if err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext())
	}

	return msg, err
}

func (ctx *MessageContext) buildEmbed(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) *embed.Embed {
	return utils.BuildEmbed(ctx.worker, ctx.GuildId(), colour, title, content, fields, ctx.premium > premium.None, format...)
}

func (ctx *MessageContext) buildEmbedRaw(colour utils.Colour, title, content string, fields ...embed.EmbedField) *embed.Embed {
	return utils.BuildEmbedRaw(ctx.worker, colour, title, content, fields, ctx.premium > premium.None)
}

func (ctx *MessageContext) Reply(colour utils.Colour, title string, content translations.MessageId, format ...interface{}) {
	if msg, err := ctx.reply(ctx.buildEmbed(colour, title, content, nil, format...)); err == nil {
		utils.DeleteAfter(ctx.worker, msg.ChannelId, msg.Id, utils.DeleteAfterSeconds)
	}
}

func (ctx *MessageContext) ReplyWithEmbed(embed *embed.Embed) {
	if msg, err := ctx.reply(embed); err == nil {
		utils.DeleteAfter(ctx.worker, msg.ChannelId, msg.Id, utils.DeleteAfterSeconds)
	}
}

func (ctx *MessageContext) ReplyWithEmbedPermanent(embed *embed.Embed) (message.Message, error) {
	return ctx.reply(embed)
}

func (ctx *MessageContext) ReplyPermanent(colour utils.Colour, title string, content translations.MessageId, format ...interface{}) {
	_, _ = ctx.reply(ctx.buildEmbed(colour, title, content, nil, format...))
}

func (ctx *MessageContext) ReplyWithFields(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) {
	if msg, err := ctx.reply(ctx.buildEmbed(colour, title, content, fields, format...)); err == nil {
		utils.DeleteAfter(ctx.worker, msg.ChannelId, msg.Id, utils.DeleteAfterSeconds)
	}
}

func (ctx *MessageContext) ReplyWithFieldsPermanent(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{}) {
	_, _ = ctx.reply(ctx.buildEmbed(colour, title, content, fields, format...))
}

func (ctx *MessageContext) ReplyRaw(colour utils.Colour, title, content string) {
	if msg, err := ctx.reply(ctx.buildEmbedRaw(colour, title, content)); err == nil {
		utils.DeleteAfter(ctx.worker, msg.ChannelId, msg.Id, utils.DeleteAfterSeconds)
	}
}

func (ctx *MessageContext) ReplyRawPermanent(colour utils.Colour, title, content string) {
	_, _ = ctx.reply(ctx.buildEmbedRaw(colour, title, content))
}

func (ctx *MessageContext) ReplyPlain(content string) {
	if msg, err := ctx.replyRaw(content); err == nil {
		utils.DeleteAfter(ctx.worker, msg.ChannelId, msg.Id, utils.DeleteAfterSeconds)
	}
}

func (ctx *MessageContext) ReplyPlainPermanent(content string) {
	_, _ = ctx.replyRaw(content)
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
	_, _ = ctx.reply(embed)
}

func (ctx *MessageContext) HandleWarning(err error) {
	sentry.LogWithContext(err, ctx.ToErrorContext())

	embed := ctx.buildEmbedRaw(utils.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()))
	_, _ = ctx.reply(embed)
}

func (ctx *MessageContext) Guild() (guild.Guild, error) {
	return ctx.Worker().GetGuild(ctx.GuildId())
}

func (ctx *MessageContext) Member() (member.Member, error) {
	return ctx.Worker().GetGuildMember(ctx.GuildId(), ctx.UserId())
}

func (ctx *MessageContext) User() (user.User, error) {
	return ctx.Worker().GetUser(ctx.UserId())
}
