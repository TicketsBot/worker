package setup

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/setup"
	"github.com/TicketsBot/worker/bot/utils"
)

type EasySetupCommand struct{}

func (EasySetupCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "ez",
		Description:     translations.HelpSetupEasy,
		Aliases:         []string{"easy"},
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
		MessageOnly:     true,
	}
}

func (c EasySetupCommand) GetExecutor() interface{} {
	return c.Execute
}

func (EasySetupCommand) Execute(ctx registry.CommandContext) {
	u := setup.FromContext(ctx)

	if u.InSetup() {
		ctx.Reject()
		ctx.Reply(utils.Red, "Error", translations.MessageAlreadyInSetup, utils.DEFAULT_PREFIX)
	} else {
		ctx.Accept()

		u.Next()
		state := u.GetState()
		stage := state.GetStage()
		if stage != nil {
			// Psuedo-premium
			// TODO: Delete after
			ctx.ReplyPermanent(utils.Green, "Setup", (*stage).Prompt())
		}
	}
}
