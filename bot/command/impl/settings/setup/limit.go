package setup

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"strconv"
)

type LimitSetupCommand struct{}

func (LimitSetupCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "limit",
		Description:     translations.HelpSetup,
		Aliases:         []string{"ticketlimit", "max", "maximum"},
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
	}
}

func (LimitSetupCommand) Execute(ctx command.CommandContext) {
	if len(ctx.Args) == 0 {
		ctx.SendEmbed(utils.Red, "Setup", translations.SetupLimitInvalid)
		ctx.ReactWithCross()
		return
	}

	limit, err := strconv.Atoi(ctx.Args[0])
	if err != nil || limit < 1 || limit > 10 {
		ctx.SendEmbed(utils.Red, "Setup", translations.SetupLimitInvalid)
		ctx.ReactWithCross()
		return
	}

	if err := dbclient.Client.TicketLimit.Set(ctx.GuildId, uint8(limit)); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.SendEmbed(utils.Green, "Setup", translations.SetupLimitComplete, limit)
	ctx.ReactWithCheck()
}
