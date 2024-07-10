package context

import (
	"context"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
)

type AutoCloseContext struct {
	context.Context
	*Replyable
	*StateCache
	worker                     *worker.Context
	guildId, channelId, userId uint64
	premium                    premium.PremiumTier
}

var _ registry.CommandContext = (*AutoCloseContext)(nil)

func NewAutoCloseContext(
	ctx context.Context,
	worker *worker.Context,
	guildId, channelId, userId uint64,
	premium premium.PremiumTier,
) *AutoCloseContext {
	c := AutoCloseContext{
		Context:   ctx,
		worker:    worker,
		guildId:   guildId,
		channelId: channelId,
		userId:    userId,
		premium:   premium,
	}

	c.Replyable = NewReplyable(&c)
	c.StateCache = NewStateCache(&c)
	return &c
}

func (c *AutoCloseContext) Worker() *worker.Context {
	return c.worker
}

func (c *AutoCloseContext) GuildId() uint64 {
	return c.guildId
}

func (c *AutoCloseContext) ChannelId() uint64 {
	return c.channelId
}

func (c *AutoCloseContext) UserId() uint64 {
	return c.userId
}

// TODO: Could this be dangerous? Don't think so, since this context is only used for closing
func (c *AutoCloseContext) UserPermissionLevel(ctx context.Context) (permcache.PermissionLevel, error) {
	return permcache.Admin, nil
}

func (c *AutoCloseContext) PremiumTier() premium.PremiumTier {
	return c.premium
}

func (c *AutoCloseContext) IsInteraction() bool {
	return true
}

func (c *AutoCloseContext) Source() registry.Source {
	return registry.SourceAutoClose
}

func (c *AutoCloseContext) ToErrorContext() errorcontext.WorkerErrorContext {
	return errorcontext.WorkerErrorContext{
		Guild:   c.guildId,
		User:    c.userId,
		Channel: c.channelId,
	}
}

func (c *AutoCloseContext) openDm() (uint64, bool) {
	return 0, false
}

func (c *AutoCloseContext) ReplyWith(response command.MessageResponse) (message.Message, error) {
	return message.Message{}, nil
}

func (c *AutoCloseContext) Channel() (channel.PartialChannel, error) {
	ch, err := c.Worker().GetChannel(c.channelId)
	if err != nil {
		return channel.PartialChannel{}, err
	}

	return ch.ToPartialChannel(), nil
}

func (c *AutoCloseContext) Guild() (guild.Guild, error) {
	return c.Worker().GetGuild(c.guildId)
}

func (c *AutoCloseContext) Member() (member.Member, error) {
	return c.Worker().GetGuildMember(c.guildId, c.userId)
}

func (c *AutoCloseContext) User() (user.User, error) {
	return c.Worker().GetUser(c.UserId())
}

func (c *AutoCloseContext) IsBlacklisted(ctx context.Context) (bool, error) {
	permLevel, err := c.UserPermissionLevel(ctx)
	if err != nil {
		return false, err
	}

	member, err := c.Member()
	if err != nil {
		return false, err
	}

	// if interaction.Member is nil, it does not matter, as the member's roles are not checked
	// if the command is not executed in a guild
	return utils.IsBlacklisted(ctx, c.GuildId(), c.UserId(), member, permLevel)
}
