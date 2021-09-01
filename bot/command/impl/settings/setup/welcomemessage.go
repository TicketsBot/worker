package setup

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
)

type WelcomeMessageSetupCommand struct{}

func (WelcomeMessageSetupCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "welcomemessage",
		Description:     i18n.HelpSetup,
		Type:            interaction.ApplicationCommandTypeChatInput,
		Aliases:         []string{"wm", "welcome"},
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewRequiredArgument("message", "The initial message sent in ticket channels", interaction.OptionTypeString, i18n.SetupWelcomeMessageInvalid),
		),
	}
}

func (c WelcomeMessageSetupCommand) GetExecutor() interface{} {
	return c.Execute
}

func (WelcomeMessageSetupCommand) Execute(ctx registry.CommandContext, message string) {
	if len(message) > 1024 {
		ctx.Reply(utils.Red, "Setup", i18n.SetupWelcomeMessageInvalid)
		ctx.Reject()
		return
	}

	if err := dbclient.Client.WelcomeMessages.Set(ctx.GuildId(), message); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Reply(utils.Green, "Setup", i18n.SetupWelcomeMessageComplete)
	ctx.Accept()
}
