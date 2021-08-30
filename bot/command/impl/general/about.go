package general

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
)

type AboutCommand struct {
}

func (AboutCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:             "about",
		Description:      i18n.HelpAbout,
		PermissionLevel:  permission.Everyone,
		Category:         command.General,
		MainBotOnly:      true,
		DefaultEphemeral: true,
	}
}

func (c AboutCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AboutCommand) Execute(ctx registry.CommandContext) {
	ctx.Reply(utils.Green, "About", i18n.MessageAbout)
}
