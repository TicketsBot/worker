package setup

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
)

type WelcomeMessageSetupCommand struct{}

func (WelcomeMessageSetupCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "welcomemessage",
		Description:     translations.HelpSetup,
		Aliases:         []string{"wm", "welcome"},
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewRequiredArgument("message", "The initial message sent in ticket channels", interaction.OptionTypeString, translations.SetupWelcomeMessageInvalid),
		),
	}
}

func (c WelcomeMessageSetupCommand) GetExecutor() interface{} {
	return c.Execute
}

func (WelcomeMessageSetupCommand) Execute(ctx command.CommandContext, message string) {
	if len(message) > 1024 {
		ctx.Reply(utils.Red, "Setup", translations.SetupWelcomeMessageInvalid)
		ctx.Reject()
		return
	}

	if err := dbclient.Client.WelcomeMessages.Set(ctx.GuildId(), message); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Reply(utils.Green, "Setup", translations.SetupWelcomeMessageComplete)
	ctx.Accept()
}
