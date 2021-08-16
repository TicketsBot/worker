package context

import (
	"errors"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
	"go.uber.org/atomic"
)

type SlashCommandContext struct {
	*Replyable
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
		worker:      worker,
		Interaction: interaction,
		premium:     premium,

		hasReplied: atomic.NewBool(false),
		responseCh: responseCh,
	}

	ctx.Replyable = NewReplyable(&ctx)
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
	return ctx.Interaction.Member.User.Id
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

func (ctx *SlashCommandContext) ToErrorContext() errorcontext.WorkerErrorContext {
	return errorcontext.WorkerErrorContext{
		Guild:   ctx.GuildId(),
		User:    ctx.Interaction.Member.User.Id,
		Channel: ctx.ChannelId(),
	}
}

func (ctx *SlashCommandContext) ReplyWith(response registry.MessageResponse) (message.Message, error) {
	hasReplied := ctx.hasReplied.Swap(true)

	if hasReplied {
		// TODO: Should we wait?
		msg, err := ctx.worker.ExecuteWebhook(ctx.worker.BotId, ctx.Interaction.Token, true, response.IntoWebhookBody())

		if err != nil {
			sentry.LogWithContext(err, ctx.ToErrorContext())
		}

		if msg == nil {
			return message.Message{}, errors.New("message was nil")
		} else {
			return *msg, err
		}
	} else {
		ctx.responseCh <- response.IntoApplicationCommandData()

		// todo: uhm
		return message.Message{}, nil
	}
}

func (ctx *SlashCommandContext) Accept() {}

func (ctx *SlashCommandContext) Reject() {}

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
