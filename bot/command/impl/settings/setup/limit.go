package setup

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"time"
)

type LimitSetupCommand struct{}

func (LimitSetupCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "limit",
		Description:     i18n.HelpSetup,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewRequiredArgument("limit", "The maximum amount of tickets a user can have open simultaneously", interaction.OptionTypeInteger, i18n.SetupLimitInvalid),
		),
		Timeout: time.Second * 3,
	}
}

func (c LimitSetupCommand) GetExecutor() interface{} {
	return c.Execute
}

func (LimitSetupCommand) Execute(ctx registry.CommandContext, limit int) {
	if limit < 1 || limit > 10 {
		ctx.Reply(customisation.Red, i18n.TitleSetup, i18n.SetupLimitInvalid)
		return
	}

	if err := dbclient.Client.TicketLimit.Set(ctx, ctx.GuildId(), uint8(limit)); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Reply(customisation.Green, i18n.TitleSetup, i18n.SetupLimitComplete, limit)
}
