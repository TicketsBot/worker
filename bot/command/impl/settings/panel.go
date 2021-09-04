package settings

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
)

type PanelCommand struct {
}

func (PanelCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "panel",
		Description:     i18n.HelpPanel,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
	}
}

func (c PanelCommand) GetExecutor() interface{} {
	return c.Execute
}

func (PanelCommand) Execute(ctx registry.CommandContext) {
	ctx.Reply(constants.Green, i18n.TitlePanel, i18n.MessagePanel, ctx.GuildId())
}
