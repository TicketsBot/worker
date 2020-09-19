package setup

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"strings"
)

type WelcomeMessageSetupCommand struct{}

func (WelcomeMessageSetupCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "welcomemessage",
		Description:     translations.HelpSetup,
		Aliases:         []string{"wm", "welcome"},
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
	}
}

func (WelcomeMessageSetupCommand) Execute(ctx command.CommandContext) {
	if len(ctx.Args) == 0 {
		ctx.SendEmbed(utils.Red, "Setup", translations.SetupWelcomeMessageInvalid)
		ctx.ReactWithCross()
		return
	}

	message := strings.Join(ctx.Args, " ")
	if len(message) > 1024 {
		ctx.SendEmbed(utils.Red, "Setup", translations.SetupWelcomeMessageInvalid)
		ctx.ReactWithCross()
		return
	}

	if err := dbclient.Client.WelcomeMessages.Set(ctx.GuildId, message); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.SendEmbed(utils.Green, "Setup", translations.SetupWelcomeMessageComplete)
	ctx.ReactWithCheck()
}
