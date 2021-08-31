package settings

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
)

type AutoCloseConfigureCommand struct {
}

func (AutoCloseConfigureCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "configure",
		Description:     i18n.HelpAutoCloseConfigure,
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
	}
}

func (c AutoCloseConfigureCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AutoCloseConfigureCommand) Execute(ctx registry.CommandContext) {
	ctx.Reply(utils.Green, "Autoclose", i18n.MessageAutoCloseConfigure, ctx.GuildId())
}
