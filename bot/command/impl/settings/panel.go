package settings

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/utils"
)

type PanelCommand struct {
}

func (PanelCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "panel",
		Description:     i18n.HelpPanel,
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
	}
}

func (c PanelCommand) GetExecutor() interface{} {
	return c.Execute
}

func (PanelCommand) Execute(ctx registry.CommandContext) {
	ctx.Reply(utils.Green, "Panel", i18n.MessagePanel, ctx.GuildId())
}
