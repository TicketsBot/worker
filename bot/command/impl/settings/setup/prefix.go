package setup

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
)

type PrefixSetupCommand struct{}

func (PrefixSetupCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "prefix",
		Description:     translations.HelpSetup,
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
	}
}

func (PrefixSetupCommand) Execute(ctx command.CommandContext) {
	if len(ctx.Args) == 0 || len(ctx.Args[0]) > 8 {
		ctx.SendEmbed(utils.Red, "Setup", translations.SetupPrefixInvalid)
		ctx.ReactWithCross()
		return
	}

	prefix := ctx.Args[0]
	if err := dbclient.Client.Prefix.Set(ctx.GuildId, prefix); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.SendEmbed(utils.Green, "Setup", translations.SetupPrefixComplete, prefix, prefix)
	ctx.ReactWithCheck()
}

