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
	"github.com/rxdn/gdl/objects/interaction/component"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/rest"
	"go.uber.org/atomic"
	"time"
)

type ModalContext struct {
	context.Context
	*Replyable
	*ReplyCounter
	*MessageComponentExtensions
	*StateCache
	worker          *worker.Context
	Interaction     interaction.ModalSubmitInteraction
	premium         premium.PremiumTier
	hasReplied      *atomic.Bool
	responseChannel chan button.Response
}

var _ registry.CommandContext = (*ModalContext)(nil)

func NewModalContext(
	ctx context.Context,
	worker *worker.Context,
	interaction interaction.ModalSubmitInteraction,
	premium premium.PremiumTier,
	responseChannel chan button.Response,
) *ModalContext {
	c := ModalContext{
		Context:         ctx,
		ReplyCounter:    NewReplyCounter(),
		worker:          worker,
		Interaction:     interaction,
		premium:         premium,
		hasReplied:      atomic.NewBool(false),
		responseChannel: responseChannel,
	}

	c.Replyable = NewReplyable(&c)
	c.MessageComponentExtensions = NewMessageComponentExtensions(&c, interaction.InteractionMetadata, responseChannel, c.hasReplied)
	c.StateCache = NewStateCache(&c)
	return &c
}

func (c *ModalContext) Defer() {
	c.hasReplied.Store(true)
	c.Ack()
}

func (c *ModalContext) GetInput(customId string) (string, bool) {
	for _, c := range c.Interaction.Data.Components {
		if c.Type != component.ComponentActionRow || len(c.Components) != 1 {
			continue
		}

		input := c.Components[0]
		if input.Type != component.ComponentInputText {
			continue
		}

		if input.CustomId == customId {
			return input.Value, true
		}
	}

	return "", false
}

func (c *ModalContext) Worker() *worker.Context {
	return c.worker
}

func (c *ModalContext) GuildId() uint64 {
	return c.Interaction.GuildId.Value // TODO: Null check
}

func (c *ModalContext) ChannelId() uint64 {
	return c.Interaction.ChannelId
}

func (c *ModalContext) UserId() uint64 {
	return c.InteractionUser().Id
}

func (c *ModalContext) UserPermissionLevel(ctx context.Context) (permcache.PermissionLevel, error) {
	if c.Interaction.Member == nil {
		return permcache.Everyone, errors.New("member was nil")
	}

	return permcache.GetPermissionLevel(ctx, utils.ToRetriever(c.worker), *c.Interaction.Member, c.GuildId())
}

func (c *ModalContext) PremiumTier() premium.PremiumTier {
	return c.premium
}

func (c *ModalContext) IsInteraction() bool {
	return true
}

func (c *ModalContext) Source() registry.Source {
	return registry.SourceDiscord
}

func (c *ModalContext) ToErrorContext() errorcontext.WorkerErrorContext {
	return errorcontext.WorkerErrorContext{
		Guild:   c.GuildId(),
		User:    c.UserId(),
		Channel: c.ChannelId(),
	}
}

func (c *ModalContext) ReplyWith(response command.MessageResponse) (msg message.Message, err error) {
	hasReplied := c.hasReplied.Swap(true)

	if err := c.ReplyCounter.Try(); err != nil {
		return message.Message{}, err
	}

	if !hasReplied {
		c.responseChannel <- button.ResponseMessage{
			Data: response,
		}
	} else {
		if time.Now().Sub(utils.SnowflakeToTime(c.interaction.Id)) > time.Minute*14 {
			return
		}

		msg, err = rest.CreateFollowupMessage(context.Background(), c.Interaction.Token, c.worker.RateLimiter, c.worker.BotId, response.IntoWebhookBody())
		if err != nil {
			sentry.LogWithContext(err, c.ToErrorContext())
		}
	}

	return
}

func (c *ModalContext) Channel() (channel.PartialChannel, error) {
	return c.Interaction.Channel, nil
}

func (c *ModalContext) Guild() (guild.Guild, error) {
	return c.Worker().GetGuild(c.GuildId())
}

func (c *ModalContext) Member() (member.Member, error) {
	if c.GuildId() == 0 {
		return member.Member{}, fmt.Errorf("button was not clicked in a guild")
	}

	if c.Interaction.Member != nil {
		return *c.Interaction.Member, nil
	} else {
		return c.Worker().GetGuildMember(c.GuildId(), c.UserId())
	}
}

func (c *ModalContext) InteractionMember() member.Member {
	if c.Interaction.Member != nil {
		return *c.Interaction.Member
	} else {
		sentry.ErrorWithContext(fmt.Errorf("ModalContext.InteractionMember was called when Member is nil"), c.ToErrorContext())
		return member.Member{}
	}
}

func (c *ModalContext) User() (user.User, error) {
	return c.InteractionUser(), nil
}

func (c *ModalContext) InteractionUser() user.User {
	if c.Interaction.Member != nil {
		return c.Interaction.Member.User
	} else if c.Interaction.User != nil {
		return *c.Interaction.User
	} else { // Infallible
		sentry.ErrorWithContext(fmt.Errorf("infallible: ModalContext.InteractionUser was called when User is nil"), c.ToErrorContext())
		return user.User{}
	}
}

func (c *ModalContext) IntoPanelContext() PanelContext {
	return NewPanelContext(c.Context, c.worker, c.GuildId(), c.ChannelId(), c.InteractionUser().Id, c.PremiumTier())
}

func (c *ModalContext) IsBlacklisted(ctx context.Context) (bool, error) {
	// TODO: Check user blacklist
	if c.GuildId() == 0 {
		return false, nil
	}

	permLevel, err := c.UserPermissionLevel(ctx)
	if err != nil {
		return false, err
	}

	// if interaction.Member is nil, it does not matter, as the member's roles are not checked
	// if the command is not executed in a guild
	return utils.IsBlacklisted(ctx, c.GuildId(), c.UserId(), utils.ValueOrZero(c.Interaction.Member), permLevel)
}

/// InteractionContext functions

func (c *ModalContext) InteractionMetadata() interaction.InteractionMetadata {
	return c.Interaction.InteractionMetadata
}
