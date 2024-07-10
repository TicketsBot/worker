package settings

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"time"
)

type AutoCloseConfigureCommand struct {
}

func (AutoCloseConfigureCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:             "configure",
		Description:      i18n.HelpAutoCloseConfigure,
		Type:             interaction.ApplicationCommandTypeChatInput,
		PermissionLevel:  permission.Admin,
		Category:         command.Settings,
		DefaultEphemeral: true,
		Timeout:          time.Second * 3,
	}
}

func (c AutoCloseConfigureCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AutoCloseConfigureCommand) Execute(ctx registry.CommandContext) {
	ctx.Reply(customisation.Green, i18n.TitleAutoclose, i18n.MessageAutoCloseConfigure, ctx.GuildId())
}
