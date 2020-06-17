package command

import (
	"fmt"
)

type Command interface {
	Execute(ctx CommandContext)
	Properties() Properties
}

func FormatHelp(c Command, prefix string) string {
	return fmt.Sprintf("**%s%s**: %s", prefix, c.Properties().Name, c.Properties().Description)
}