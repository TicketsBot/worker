package general

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/utils"
)

type AboutCommand struct {
}

func (AboutCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "about",
		Description:     translations.HelpAbout,
		PermissionLevel: permission.Everyone,
		Category:        command.General,
	}
}

func (AboutCommand) Execute(ctx command.CommandContext) {
	ctx.SendEmbed(utils.Green, "About", translations.MessageAbout)
}
