package impl

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/utils"
)

type PanelCommand struct {
}

func (PanelCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "panel",
		Description:     "Creates a panel to enable users to open a ticket with 1 click",
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
	}
}

func (PanelCommand) Execute(ctx command.CommandContext) {
	msg := fmt.Sprintf("Visit https://panel.ticketsbot.net/manage/%d/panels to configure a panel", ctx.GuildId)
	ctx.SendEmbed(utils.Green, "Panel", msg)
}
