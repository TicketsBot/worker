package context

import (
	"errors"
	"fmt"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/rest"
)

type ButtonContext struct {
	*Replyable
	worker      *worker.Context
	Interaction interaction.ButtonInteraction
	premium     premium.PremiumTier
	editChannel chan command.MessageResponse
}

func NewButtonContext(
	worker *worker.Context,
	interaction interaction.ButtonInteraction,
	premium premium.PremiumTier,
	editChannel chan command.MessageResponse,
) *ButtonContext {
	ctx := ButtonContext{
		worker:      worker,
		Interaction: interaction,
		premium:     premium,
		editChannel: editChannel,
	}

	ctx.Replyable = NewReplyable(&ctx)
	return &ctx
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
	return ctx.InteractionUser().Id
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

func (ctx *ButtonContext) ReplyWith(response command.MessageResponse) (message.Message, error) {
	msg, err := rest.CreateFollowupMessage(ctx.Interaction.Token, ctx.worker.RateLimiter, ctx.worker.BotId, response.IntoWebhookBody())
	if err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext())
	}

	return msg, err
}

func (ctx *ButtonContext) Edit(response command.MessageResponse) {
	ctx.editChannel <- response
}

func (ctx *ButtonContext) Accept() {}

func (ctx *ButtonContext) Reject() {}

func (ctx *ButtonContext) Guild() (guild.Guild, error) {
	return ctx.Worker().GetGuild(ctx.GuildId())
}

func (ctx *ButtonContext) Member() (member.Member, error) {
	if ctx.GuildId() == 0 {
		return member.Member{}, fmt.Errorf("button was not clicked in a guild")
	}

	if ctx.Interaction.Member != nil {
		return *ctx.Interaction.Member, nil
	} else {
		return ctx.Worker().GetGuildMember(ctx.GuildId(), ctx.UserId())
	}
}

func (ctx *ButtonContext) InteractionMember() member.Member {
	if ctx.Interaction.Member != nil {
		return *ctx.Interaction.Member
	} else {
		sentry.ErrorWithContext(fmt.Errorf("ButtonContext.InteractionMember was called when Member is nil"), ctx.ToErrorContext())
		return member.Member{}
	}
}

func (ctx *ButtonContext) User() (user.User, error) {
	return ctx.InteractionUser(), nil
}

func (ctx *ButtonContext) InteractionUser() user.User {
	if ctx.Interaction.Member != nil {
		return ctx.Interaction.Member.User
	} else if ctx.Interaction.User != nil {
		return *ctx.Interaction.User
	} else { // Infallible
		sentry.ErrorWithContext(fmt.Errorf("infallible: ButtonContext.InteractionUser was called when User is nil"), ctx.ToErrorContext())
		return user.User{}
	}
}

func (ctx *ButtonContext) IntoPanelContext() PanelContext {
	return NewPanelContext(ctx.worker, ctx.GuildId(), ctx.ChannelId(), ctx.InteractionUser().Id, ctx.PremiumTier())
}
