package setup

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
)

type LimitSetupCommand struct{}

func (LimitSetupCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "limit",
		Description:     translations.HelpSetup,
		Aliases:         []string{"ticketlimit", "max", "maximum"},
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewRequiredArgument("limit", "The maximum amount of tickets a user can have open simultaneously", interaction.OptionTypeInteger, translations.SetupLimitInvalid),
		),
	}
}

func (c LimitSetupCommand) GetExecutor() interface{} {
	return c.Execute
}

func (LimitSetupCommand) Execute(ctx command.CommandContext, limit int) {
	if limit < 1 || limit > 10 {
		ctx.Reply(utils.Red, "Setup", translations.SetupLimitInvalid)
		ctx.Reject()
		return
	}

	if err := dbclient.Client.TicketLimit.Set(ctx.GuildId(), uint8(limit)); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Reply(utils.Green, "Setup", translations.SetupLimitComplete, limit)
	ctx.Accept()
}
