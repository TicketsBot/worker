package setup

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
	"strings"
)

type PrefixSetupCommand struct{}

func (PrefixSetupCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "prefix",
		Description:     translations.HelpSetup,
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewRequiredArgument("prefix", "Characters that come before the command, i.e. t!", interaction.OptionTypeString, translations.SetupPrefixInvalid),
		),
	}
}

func (c PrefixSetupCommand) GetExecutor() interface{} {
	return c.Execute
}

func (PrefixSetupCommand) Execute(ctx command.CommandContext, prefix string) {
	if len(prefix) == 0 || len(prefix) > 8 || strings.Contains(prefix, " ") {
		ctx.Reply(utils.Red, "Setup", translations.SetupPrefixInvalid)
		ctx.Reject()
		return
	}

	if err := dbclient.Client.Prefix.Set(ctx.GuildId(), prefix); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Reply(utils.Green, "Setup", translations.SetupPrefixComplete, prefix, prefix)
	ctx.Accept()
}
