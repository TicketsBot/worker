package setup

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/setup"
	"github.com/TicketsBot/worker/bot/utils"
)

type EasySetupCommand struct{}

func (EasySetupCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "ez",
		Description:     translations.HelpSetupEasy,
		Aliases:         []string{"easy"},
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
	}
}

func (EasySetupCommand) Execute(ctx command.CommandContext) {
	u := setup.FromContext(ctx)

	if u.InSetup() {
		ctx.ReactWithCross()
		ctx.Reply(utils.Red, "Error", translations.MessageAlreadyInSetup, utils.DEFAULT_PREFIX)
	} else {
		ctx.ReactWithCheck()

		u.Next()
		state := u.GetState()
		stage := state.GetStage()
		if stage != nil {
			// Psuedo-premium
			// TODO: TRANSLATE SETUP PROMPTS
			ctx.ReplyWithDeleteAfter(utils.Green, "Setup", (*stage).Prompt(), 120)
		}
	}
}

