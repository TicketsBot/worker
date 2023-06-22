package context

import (
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
)

type MessageContext struct {
	*Replyable
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
	ctx := MessageContext{
		worker:          worker,
		Message:         message,
		Args:            args,
		premium:         premium,
		permissionLevel: permissionLevel,
	}

	ctx.Replyable = NewReplyable(&ctx)
	return ctx
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
	}
}

func (ctx *MessageContext) ReplyContext() *message.MessageReference {
	return &message.MessageReference{
		MessageId: ctx.Id,
		ChannelId: ctx.ChannelId(),
		GuildId:   ctx.GuildId(),
	}
}

func (ctx *MessageContext) Reply(colour customisation.Colour, title, content i18n.MessageId, format ...interface{}) {
	embed := ctx.Replyable.buildEmbed(colour, title, content, nil, format...)
	embed.SetFooter(ctx.GetMessage(i18n.MessageUseSlashCommands), ctx.avatarUrl())

	_, _ = ctx.ReplyWith(command.NewEphemeralEmbedMessageResponse(embed))
}

func (ctx *MessageContext) ReplyWith(response command.MessageResponse) (message.Message, error) {
	data := response.IntoCreateMessageData()
	data.MessageReference = ctx.ReplyContext()

	msg, err := ctx.worker.CreateMessageComplex(ctx.ChannelId(), data)
	if err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext())
	}

	return msg, err
}

func (ctx *MessageContext) Accept() {
	utils.ReactWithCheck(ctx.worker, ctx.ChannelId(), ctx.Id)
}

func (ctx *MessageContext) Reject() {
	utils.ReactWithCross(ctx.worker, ctx.ChannelId(), ctx.Id)
}

func (ctx *MessageContext) Channel() (channel.PartialChannel, error) {
	ch, err := ctx.Worker().GetChannel(ctx.ChannelId())
	if err != nil {
		return channel.PartialChannel{}, err
	}

	return ch.ToPartialChannel(), nil
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

func (ctx *MessageContext) avatarUrl() string {
	self, err := ctx.Worker().Self()
	if err == nil {
		return self.AvatarUrl(256)
	} else {
		return "https://ticketsbot.net/assets/img/logo.png"
	}
}

func (ctx *MessageContext) IsBlacklisted() (bool, error) {
	permLevel, err := ctx.UserPermissionLevel()
	if err != nil {
		return false, err
	}

	member, err := ctx.Member()
	if err != nil {
		return false, err
	}

	// if interaction.Member is nil, it does not matter, as the member's roles are not checked
	// if the command is not executed in a guild
	return utils.IsBlacklisted(ctx.GuildId(), ctx.UserId(), member, permLevel)
}
