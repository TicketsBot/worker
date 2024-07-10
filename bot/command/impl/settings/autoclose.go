package settings

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
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
	// Can't call a parent command
}
