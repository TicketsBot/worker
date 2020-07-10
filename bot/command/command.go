package command

import (
	"fmt"
	"github.com/TicketsBot/worker/bot/i18n"
)

type Command interface {
	Execute(ctx CommandContext)
	Properties() Properties
}

func FormatHelp(c Command, guildId uint64, prefix string) string {
	description := i18n.GetMessageFromGuild(guildId, c.Properties().Description)
	return fmt.Sprintf("**%s%s**: %s", prefix, c.Properties().Name, description)
}