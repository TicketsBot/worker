package setup

import (
	"fmt"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
)

type SetupUser struct {
	Guild   uint64
	User    uint64
	Channel uint64
	Worker  *worker.Context
}

func (s *SetupUser) ToString() string {
	return fmt.Sprintf("%d:%d:%d", s.Guild, s.User, s.Channel)
}

func FromContext(ctx command.CommandContext) SetupUser {
	return SetupUser{
		Guild:   ctx.GuildId(),
		User:    ctx.UserId(),
		Channel: ctx.ChannelId(),
		Worker:  ctx.Worker(),
	}
}
