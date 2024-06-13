package registry

import (
	"fmt"
	"github.com/TicketsBot/worker/i18n"
	"strings"
)

type Command interface {
	GetExecutor() interface{}
	//Execute(ctx CommandContext)
	Properties() Properties
}

func FormatHelp(c Command, guildId uint64, commandId *uint64) string {
	description := i18n.GetMessageFromGuild(guildId, c.Properties().Description)

	if commandId == nil {
		var args []string
		for _, arg := range c.Properties().Arguments {
			if arg.Required {
				args = append(args, fmt.Sprintf("[%s] ", arg.Name))
			} else {
				args = append(args, fmt.Sprintf("<%s> ", arg.Name))
			}
		}

		var argsJoined string
		if len(args) > 0 {
			argsJoined = " " + strings.Join(args, " ") // Separate between command and first arg
		}

		return fmt.Sprintf("**/%s%s**: %s", c.Properties().Name, argsJoined, description)
	} else {
		return fmt.Sprintf("</%s:%d>: %s", c.Properties().Name, *commandId, description)
	}
}
