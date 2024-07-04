package context

import (
	"context"
	"errors"
	"fmt"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/button"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/rest"
	"go.uber.org/atomic"
)

type ButtonContext struct {
	*Replyable
	*ReplyCounter
	*MessageComponentExtensions
	*StateCache
	worker          *worker.Context
	Interaction     interaction.MessageComponentInteraction
	InteractionData interaction.ButtonInteractionData
	premium         premium.PremiumTier
	hasReplied      *atomic.Bool
	responseChannel chan button.Response
}

func NewButtonContext(
	worker *worker.Context,
	interaction interaction.MessageComponentInteraction,
	premium premium.PremiumTier,
	responseChannel chan button.Response,
) *ButtonContext {
	ctx := ButtonContext{
		ReplyCounter:    NewReplyCounter(),
		worker:          worker,
		Interaction:     interaction,
		InteractionData: interaction.Data.AsButton(),
		premium:         premium,
		hasReplied:      atomic.NewBool(false),
		responseChannel: responseChannel,
	}

	ctx.Replyable = NewReplyable(&ctx)
	ctx.MessageComponentExtensions = NewMessageComponentExtensions(&ctx, interaction.InteractionMetadata, responseChannel, ctx.hasReplied)
	ctx.StateCache = NewStateCache(&ctx)
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

func (ctx *ButtonContext) Source() registry.Source {
	return registry.SourceDiscord
}

func (ctx *ButtonContext) ToErrorContext() errorcontext.WorkerErrorContext {
	return errorcontext.WorkerErrorContext{
		Guild:   ctx.GuildId(),
		User:    ctx.UserId(),
		Channel: ctx.ChannelId(),
	}
}

func (ctx *ButtonContext) ReplyWith(response command.MessageResponse) (msg message.Message, err error) {
	hasReplied := ctx.hasReplied.Swap(true)

	if err := ctx.ReplyCounter.Try(); err != nil {
		return message.Message{}, err
	}

	if !hasReplied {
		ctx.responseChannel <- button.ResponseMessage{
			Data: response,
		}
	} else {
		msg, err = rest.CreateFollowupMessage(context.Background(), ctx.Interaction.Token, ctx.worker.RateLimiter, ctx.worker.BotId, response.IntoWebhookBody())
		if err != nil {
			sentry.LogWithContext(err, ctx.ToErrorContext())
		}
	}

	return
}

func (ctx *ButtonContext) Channel() (channel.PartialChannel, error) {
	return ctx.Interaction.Channel, nil
}

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

func (ctx *ButtonContext) IsBlacklisted() (bool, error) {
	// TODO: Check user blacklist
	if ctx.GuildId() == 0 {
		return false, nil
	}

	permLevel, err := ctx.UserPermissionLevel()
	if err != nil {
		return false, err
	}

	// if interaction.Member is nil, it does not matter, as the member's roles are not checked
	// if the command is not executed in a guild
	return utils.IsBlacklisted(ctx.GuildId(), ctx.UserId(), utils.ValueOrZero(ctx.Interaction.Member), permLevel)
}

/// InteractionContext functions

func (ctx *ButtonContext) InteractionMetadata() interaction.InteractionMetadata {
	return ctx.Interaction.InteractionMetadata
}
