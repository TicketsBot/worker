package setup

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"strings"
)

type PrefixSetupCommand struct{}

func (PrefixSetupCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "prefix",
		Description:     i18n.HelpSetup,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewRequiredArgument("prefix", "Characters that come before the command, i.e. t!", interaction.OptionTypeString, i18n.SetupPrefixInvalid),
		),
	}
}

func (c PrefixSetupCommand) GetExecutor() interface{} {
	return c.Execute
}

func (PrefixSetupCommand) Execute(ctx registry.CommandContext, prefix string) {
	if len(prefix) == 0 || len(prefix) > 8 || strings.Contains(prefix, " ") {
		ctx.Reply(utils.Red, "Setup", i18n.SetupPrefixInvalid)
		ctx.Reject()
		return
	}

	if err := dbclient.Client.Prefix.Set(ctx.GuildId(), prefix); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Reply(utils.Green, "Setup", i18n.SetupPrefixComplete, prefix, prefix)
	ctx.Accept()
}
