package context

import (
	"errors"
	"fmt"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/button"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/rest"
	"go.uber.org/atomic"
)

type ModalContext struct {
	*Replyable
	*MessageComponentExtensions
	worker          *worker.Context
	Interaction     interaction.ModalSubmitInteraction
	premium         premium.PremiumTier
	hasReplied      *atomic.Bool
	responseChannel chan button.Response
}

func NewModalContext(
	worker *worker.Context,
	interaction interaction.ModalSubmitInteraction,
	premium premium.PremiumTier,
	responseChannel chan button.Response,
) *ModalContext {
	ctx := ModalContext{
		worker:          worker,
		Interaction:     interaction,
		premium:         premium,
		hasReplied:      atomic.NewBool(false),
		responseChannel: responseChannel,
	}

	ctx.Replyable = NewReplyable(&ctx)
	ctx.MessageComponentExtensions = NewMessageComponentExtensions(&ctx, responseChannel)
	return &ctx
}

func (ctx *ModalContext) Worker() *worker.Context {
	return ctx.worker
}

func (ctx *ModalContext) GuildId() uint64 {
	return ctx.Interaction.GuildId.Value // TODO: Null check
}

func (ctx *ModalContext) ChannelId() uint64 {
	return ctx.Interaction.ChannelId
}

func (ctx *ModalContext) UserId() uint64 {
	return ctx.InteractionUser().Id
}

func (ctx *ModalContext) UserPermissionLevel() (permcache.PermissionLevel, error) {
	if ctx.Interaction.Member == nil {
		return permcache.Everyone, errors.New("member was nil")
	}

	return permcache.GetPermissionLevel(utils.ToRetriever(ctx.worker), *ctx.Interaction.Member, ctx.GuildId())
}

func (ctx *ModalContext) PremiumTier() premium.PremiumTier {
	return ctx.premium
}

func (ctx *ModalContext) IsInteraction() bool {
	return true
}

func (ctx *ModalContext) ToErrorContext() errorcontext.WorkerErrorContext {
	return errorcontext.WorkerErrorContext{
		Guild:   ctx.GuildId(),
		User:    ctx.UserId(),
		Channel: ctx.ChannelId(),
	}
}

func (ctx *ModalContext) ReplyWith(response command.MessageResponse) (msg message.Message, err error) {
	hasReplied := ctx.hasReplied.Swap(true)

	if !hasReplied {
		ctx.responseChannel <- button.ResponseMessage{
			Data: response,
		}
	} else {
		msg, err = rest.CreateFollowupMessage(ctx.Interaction.Token, ctx.worker.RateLimiter, ctx.worker.BotId, response.IntoWebhookBody())
		if err != nil {
			sentry.LogWithContext(err, ctx.ToErrorContext())
		}
	}

	return
}

func (ctx *ModalContext) Edit(data command.MessageResponse) {
	hasReplied := ctx.hasReplied.Swap(true)

	if !hasReplied {
		ctx.responseChannel <- button.ResponseEdit{
			Data: data,
		}
	} else {
		_, err := rest.EditOriginalInteractionResponse(ctx.Interaction.Token, ctx.worker.RateLimiter, ctx.worker.BotId, data.IntoWebhookEditBody())
		if err != nil {
			sentry.LogWithContext(err, ctx.ToErrorContext())
		}
	}

	return
}

func (ctx *ModalContext) Accept() {}

func (ctx *ModalContext) Reject() {}

func (ctx *ModalContext) Guild() (guild.Guild, error) {
	return ctx.Worker().GetGuild(ctx.GuildId())
}

func (ctx *ModalContext) Member() (member.Member, error) {
	if ctx.GuildId() == 0 {
		return member.Member{}, fmt.Errorf("button was not clicked in a guild")
	}

	if ctx.Interaction.Member != nil {
		return *ctx.Interaction.Member, nil
	} else {
		return ctx.Worker().GetGuildMember(ctx.GuildId(), ctx.UserId())
	}
}

func (ctx *ModalContext) InteractionMember() member.Member {
	if ctx.Interaction.Member != nil {
		return *ctx.Interaction.Member
	} else {
		sentry.ErrorWithContext(fmt.Errorf("ModalContext.InteractionMember was called when Member is nil"), ctx.ToErrorContext())
		return member.Member{}
	}
}

func (ctx *ModalContext) User() (user.User, error) {
	return ctx.InteractionUser(), nil
}

func (ctx *ModalContext) InteractionUser() user.User {
	if ctx.Interaction.Member != nil {
		return ctx.Interaction.Member.User
	} else if ctx.Interaction.User != nil {
		return *ctx.Interaction.User
	} else { // Infallible
		sentry.ErrorWithContext(fmt.Errorf("infallible: ModalContext.InteractionUser was called when User is nil"), ctx.ToErrorContext())
		return user.User{}
	}
}

func (ctx *ModalContext) IntoPanelContext() PanelContext {
	return NewPanelContext(ctx.worker, ctx.GuildId(), ctx.ChannelId(), ctx.InteractionUser().Id, ctx.PremiumTier())
}
