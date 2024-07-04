package context

import (
	"context"
	"errors"
	"fmt"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
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

type SlashCommandContext struct {
	*Replyable
	*ReplyCounter
	*StateCache
	InteractionExtension
	worker      *worker.Context
	Interaction interaction.ApplicationCommandInteraction
	premium     premium.PremiumTier

	hasReplied *atomic.Bool
	responseCh chan interaction.ApplicationCommandCallbackData
}

func NewSlashCommandContext(
	worker *worker.Context,
	interaction interaction.ApplicationCommandInteraction,
	premium premium.PremiumTier,
	responseCh chan interaction.ApplicationCommandCallbackData,
) SlashCommandContext {
	ctx := SlashCommandContext{
		ReplyCounter: NewReplyCounter(),

		InteractionExtension: NewInteractionExtension(interaction),

		worker:      worker,
		Interaction: interaction,
		premium:     premium,

		hasReplied: atomic.NewBool(false),
		responseCh: responseCh,
	}

	ctx.Replyable = NewReplyable(&ctx)
	ctx.StateCache = NewStateCache(&ctx)
	return ctx
}

func (ctx *SlashCommandContext) Worker() *worker.Context {
	return ctx.worker
}

func (ctx *SlashCommandContext) GuildId() uint64 {
	return ctx.Interaction.GuildId.Value // TODO: Null check
}

func (ctx *SlashCommandContext) ChannelId() uint64 {
	return ctx.Interaction.ChannelId
}

func (ctx *SlashCommandContext) UserId() uint64 {
	if ctx.Interaction.Member != nil {
		return ctx.Interaction.Member.User.Id
	} else if ctx.Interaction.User != nil {
		return ctx.Interaction.User.Id
	} else {
		sentry.ErrorWithContext(fmt.Errorf("infallible: interaction.member and interaction.user are both null"), ctx.ToErrorContext())
		return 0
	}
}

func (ctx *SlashCommandContext) UserPermissionLevel() (permcache.PermissionLevel, error) {
	if ctx.Interaction.Member == nil {
		return permcache.Everyone, errors.New("member was nil")
	}

	return permcache.GetPermissionLevel(utils.ToRetriever(ctx.worker), *ctx.Interaction.Member, ctx.GuildId())
}

func (ctx *SlashCommandContext) PremiumTier() premium.PremiumTier {
	return ctx.premium
}

func (ctx *SlashCommandContext) IsInteraction() bool {
	return true
}

func (ctx *SlashCommandContext) Source() registry.Source {
	return registry.SourceDiscord
}

func (ctx *SlashCommandContext) ToErrorContext() errorcontext.WorkerErrorContext {
	return errorcontext.WorkerErrorContext{
		Guild:   ctx.GuildId(),
		User:    ctx.Interaction.Member.User.Id,
		Channel: ctx.ChannelId(),
	}
}

func (ctx *SlashCommandContext) ReplyWith(response command.MessageResponse) (message.Message, error) {
	hasReplied := ctx.hasReplied.Swap(true)

	if err := ctx.ReplyCounter.Try(); err != nil {
		return message.Message{}, err
	}

	if hasReplied {
		msg, err := rest.EditOriginalInteractionResponse(context.Background(), ctx.Interaction.Token, ctx.worker.RateLimiter, ctx.worker.BotId, response.IntoWebhookEditBody())

		if err != nil {
			sentry.LogWithContext(err, ctx.ToErrorContext())
		}

		return msg, err
	} else {
		ctx.responseCh <- response.IntoApplicationCommandData()

		// todo: uhm
		return message.Message{}, nil
	}
}

func (ctx *SlashCommandContext) Channel() (channel.PartialChannel, error) {
	return ctx.Interaction.Channel, nil
}

func (ctx *SlashCommandContext) Guild() (guild.Guild, error) {
	return ctx.Worker().GetGuild(ctx.GuildId())
}

func (ctx *SlashCommandContext) Member() (member.Member, error) {
	if ctx.Interaction.Member != nil {
		return *ctx.Interaction.Member, nil
	} else {
		return ctx.Worker().GetGuildMember(ctx.GuildId(), ctx.UserId())
	}
}

func (ctx *SlashCommandContext) User() (user.User, error) {
	return ctx.Worker().GetUser(ctx.UserId())
}

func (ctx *SlashCommandContext) IsBlacklisted() (bool, error) {
	permLevel, err := ctx.UserPermissionLevel()
	if err != nil {
		return false, err
	}

	// if interaction.Member is nil, it does not matter, as the member's roles are not checked
	// if the command is not executed in a guild
	return utils.IsBlacklisted(ctx.GuildId(), ctx.UserId(), utils.ValueOrZero(ctx.Interaction.Member), permLevel)
}

/// InteractionContext functions

func (ctx *SlashCommandContext) InteractionMetadata() interaction.InteractionMetadata {
	return ctx.Interaction.InteractionMetadata
}
