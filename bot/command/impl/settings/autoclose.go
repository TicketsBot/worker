package settings

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"strings"
)

type AutoCloseCommand struct {
}

func (AutoCloseCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "autoclose",
		Description:     i18n.HelpAutoClose,
		PermissionLevel: permission.Support,
		Category:        command.Settings,
		Children: []registry.Command{
			AutoCloseConfigureCommand{},
			AutoCloseExcludeCommand{},
		},
	}
}

func (c AutoCloseCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AutoCloseCommand) Execute(ctx registry.CommandContext) {
	msg := "Select a subcommand:\n"

	children := AutoCloseCommand{}.Properties().Children
	for _, child := range children {
		msg += fmt.Sprintf("`/autoclose %s` - %s\n", child.Properties().Name, i18n.GetMessageFromGuild(ctx.GuildId(), child.Properties().Description))
	}

	msg = strings.TrimSuffix(msg, "\n")

	ctx.ReplyRaw(utils.Red, "Error", msg)
}
