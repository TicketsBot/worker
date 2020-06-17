package general

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/utils"
)

type AboutCommand struct {
}

func (AboutCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "about",
		Description:     "Tells you information about the bot",
		PermissionLevel: permission.Everyone,
		Category:        command.General,
	}
}

func (AboutCommand) Execute(ctx command.CommandContext) {
	ctx.SendEmbed(utils.Green, "About", utils.ABOUT_MESSAGE)
}
