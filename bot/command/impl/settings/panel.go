package settings

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/utils"
)

type PanelCommand struct {
}

func (PanelCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "panel",
		Description:     translations.HelpPanel,
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
	}
}

func (PanelCommand) Execute(ctx command.CommandContext) {
	ctx.SendEmbed(utils.Green, "Panel", translations.MessagePanel, ctx.GuildId)
}
