package settings

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"strings"
)

type AutoCloseCommand struct {
}

func (AutoCloseCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "autoclose",
		Description:     i18n.HelpAutoClose,
		Type:            interaction.ApplicationCommandTypeChatInput,
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

	ctx.ReplyRaw(constants.Red, "Error", msg)
}
