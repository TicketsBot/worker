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
	"go.uber.org/atomic"
)

type ButtonContext struct {
	context.Context
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

var _ registry.CommandContext = (*ButtonContext)(nil)

func NewButtonContext(
	ctx context.Context,
	worker *worker.Context,
	interaction interaction.MessageComponentInteraction,
	premium premium.PremiumTier,
	responseChannel chan button.Response,
) *ButtonContext {
	c := ButtonContext{
		Context:         ctx,
		ReplyCounter:    NewReplyCounter(),
		worker:          worker,
		Interaction:     interaction,
		InteractionData: interaction.Data.AsButton(),
		premium:         premium,
		hasReplied:      atomic.NewBool(false),
		responseChannel: responseChannel,
	}

	c.Replyable = NewReplyable(&c)
	c.MessageComponentExtensions = NewMessageComponentExtensions(&c, interaction.InteractionMetadata, responseChannel, c.hasReplied)
	c.StateCache = NewStateCache(&c)
	return &c
}

func (c *ButtonContext) Worker() *worker.Context {
	return c.worker
}

func (c *ButtonContext) GuildId() uint64 {
	return c.Interaction.GuildId.Value // TODO: Null check
}

func (c *ButtonContext) ChannelId() uint64 {
	return c.Interaction.ChannelId
}

func (c *ButtonContext) UserId() uint64 {
	return c.InteractionUser().Id
}

func (c *ButtonContext) UserPermissionLevel(ctx context.Context) (permcache.PermissionLevel, error) {
	if c.Interaction.Member == nil {
		return permcache.Everyone, errors.New("member was nil")
	}

	return permcache.GetPermissionLevel(ctx, utils.ToRetriever(c.worker), *c.Interaction.Member, c.GuildId())
}

func (c *ButtonContext) PremiumTier() premium.PremiumTier {
	return c.premium
}

func (c *ButtonContext) IsInteraction() bool {
	return true
}

func (c *ButtonContext) Source() registry.Source {
	return registry.SourceDiscord
}

func (c *ButtonContext) ToErrorContext() errorcontext.WorkerErrorContext {
	return errorcontext.WorkerErrorContext{
		Guild:   c.GuildId(),
		User:    c.UserId(),
		Channel: c.ChannelId(),
	}
}

func (c *ButtonContext) ReplyWith(response command.MessageResponse) (msg message.Message, err error) {
	//hasReplied := c.hasReplied.Swap(true)

	if err := c.ReplyCounter.Try(); err != nil {
		return message.Message{}, err
	}

	c.responseChannel <- button.ResponseMessage{
		Data: response,
	}

	/*
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
	*/

	return
}

func (c *ButtonContext) Channel() (channel.PartialChannel, error) {
	return c.Interaction.Channel, nil
}

func (c *ButtonContext) Guild() (guild.Guild, error) {
	return c.Worker().GetGuild(c.GuildId())
}

func (c *ButtonContext) Member() (member.Member, error) {
	if c.GuildId() == 0 {
		return member.Member{}, fmt.Errorf("button was not clicked in a guild")
	}

	if c.Interaction.Member != nil {
		return *c.Interaction.Member, nil
	} else {
		return c.Worker().GetGuildMember(c.GuildId(), c.UserId())
	}
}

func (c *ButtonContext) InteractionMember() member.Member {
	if c.Interaction.Member != nil {
		return *c.Interaction.Member
	} else {
		sentry.ErrorWithContext(fmt.Errorf("ButtonContext.InteractionMember was called when Member is nil"), c.ToErrorContext())
		return member.Member{}
	}
}

func (c *ButtonContext) User() (user.User, error) {
	return c.InteractionUser(), nil
}

func (c *ButtonContext) InteractionUser() user.User {
	if c.Interaction.Member != nil {
		return c.Interaction.Member.User
	} else if c.Interaction.User != nil {
		return *c.Interaction.User
	} else { // Infallible
		sentry.ErrorWithContext(fmt.Errorf("infallible: ButtonContext.InteractionUser was called when User is nil"), c.ToErrorContext())
		return user.User{}
	}
}

func (c *ButtonContext) IntoPanelContext() PanelContext {
	return NewPanelContext(c.Context, c.worker, c.GuildId(), c.ChannelId(), c.InteractionUser().Id, c.PremiumTier())
}

func (c *ButtonContext) IsBlacklisted(ctx context.Context) (bool, error) {
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

func (c *ButtonContext) InteractionMetadata() interaction.InteractionMetadata {
	return c.Interaction.InteractionMetadata
}
