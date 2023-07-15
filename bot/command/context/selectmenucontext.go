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
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/rest"
	"go.uber.org/atomic"
)

type SelectMenuContext struct {
	*Replyable
	*MessageComponentExtensions
	*StateCache
	worker          *worker.Context
	Interaction     interaction.MessageComponentInteraction
	InteractionData interaction.SelectMenuInteractionData
	premium         premium.PremiumTier
	hasReplied      *atomic.Bool
	responseChannel chan button.Response
}

func NewSelectMenuContext(
	worker *worker.Context,
	interaction interaction.MessageComponentInteraction,
	premium premium.PremiumTier,
	responseChannel chan button.Response,
) *SelectMenuContext {
	ctx := SelectMenuContext{
		worker:          worker,
		Interaction:     interaction,
		InteractionData: interaction.Data.AsSelectMenu(),
		premium:         premium,
		hasReplied:      atomic.NewBool(false),
		responseChannel: responseChannel,
	}

	ctx.Replyable = NewReplyable(&ctx)
	ctx.MessageComponentExtensions = NewMessageComponentExtensions(&ctx, interaction.InteractionMetadata, responseChannel, ctx.hasReplied)
	ctx.StateCache = NewStateCache(&ctx)
	return &ctx
}

func (ctx *SelectMenuContext) Worker() *worker.Context {
	return ctx.worker
}

func (ctx *SelectMenuContext) GuildId() uint64 {
	return ctx.Interaction.GuildId.Value // TODO: Null check
}

func (ctx *SelectMenuContext) ChannelId() uint64 {
	return ctx.Interaction.ChannelId
}

func (ctx *SelectMenuContext) UserId() uint64 {
	return ctx.InteractionUser().Id
}

func (ctx *SelectMenuContext) UserPermissionLevel() (permcache.PermissionLevel, error) {
	if ctx.Interaction.Member == nil {
		return permcache.Everyone, errors.New("member was nil")
	}

	return permcache.GetPermissionLevel(utils.ToRetriever(ctx.worker), *ctx.Interaction.Member, ctx.GuildId())
}

func (ctx *SelectMenuContext) PremiumTier() premium.PremiumTier {
	return ctx.premium
}

func (ctx *SelectMenuContext) IsInteraction() bool {
	return true
}

func (ctx *SelectMenuContext) ToErrorContext() errorcontext.WorkerErrorContext {
	return errorcontext.WorkerErrorContext{
		Guild:   ctx.GuildId(),
		User:    ctx.UserId(),
		Channel: ctx.ChannelId(),
	}
}

func (ctx *SelectMenuContext) ReplyWith(response command.MessageResponse) (msg message.Message, err error) {
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

func (ctx *SelectMenuContext) Channel() (channel.PartialChannel, error) {
	return ctx.Interaction.Channel, nil
}

func (ctx *SelectMenuContext) Guild() (guild.Guild, error) {
	return ctx.Worker().GetGuild(ctx.GuildId())
}

func (ctx *SelectMenuContext) Member() (member.Member, error) {
	if ctx.GuildId() == 0 {
		return member.Member{}, fmt.Errorf("button was not clicked in a guild")
	}

	if ctx.Interaction.Member != nil {
		return *ctx.Interaction.Member, nil
	} else {
		return ctx.Worker().GetGuildMember(ctx.GuildId(), ctx.UserId())
	}
}

func (ctx *SelectMenuContext) InteractionMember() member.Member {
	if ctx.Interaction.Member != nil {
		return *ctx.Interaction.Member
	} else {
		sentry.ErrorWithContext(fmt.Errorf("SelectMenuContext.InteractionMember was called when Member is nil"), ctx.ToErrorContext())
		return member.Member{}
	}
}

func (ctx *SelectMenuContext) User() (user.User, error) {
	return ctx.InteractionUser(), nil
}

func (ctx *SelectMenuContext) InteractionUser() user.User {
	if ctx.Interaction.Member != nil {
		return ctx.Interaction.Member.User
	} else if ctx.Interaction.User != nil {
		return *ctx.Interaction.User
	} else { // Infallible
		sentry.ErrorWithContext(fmt.Errorf("infallible: SelectMenuContext.InteractionUser was called when User is nil"), ctx.ToErrorContext())
		return user.User{}
	}
}

func (ctx *SelectMenuContext) IntoPanelContext() PanelContext {
	return NewPanelContext(ctx.worker, ctx.GuildId(), ctx.ChannelId(), ctx.InteractionUser().Id, ctx.PremiumTier())
}

func (ctx *SelectMenuContext) IsBlacklisted() (bool, error) {
	permLevel, err := ctx.UserPermissionLevel()
	if err != nil {
		return false, err
	}

	// if interaction.Member is nil, it does not matter, as the member's roles are not checked
	// if the command is not executed in a guild
	return utils.IsBlacklisted(ctx.GuildId(), ctx.UserId(), utils.ValueOrZero(ctx.Interaction.Member), permLevel)
}

/// InteractionContext functions

func (ctx *SelectMenuContext) InteractionMetadata() interaction.InteractionMetadata {
	return ctx.Interaction.InteractionMetadata
}
