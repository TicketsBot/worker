package settings

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/setup"
	"github.com/TicketsBot/worker/bot/utils"
)

type SetupCommand struct {
}

func (SetupCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "setup",
		Description:     translations.HelpSetup,
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
	}
}

func (SetupCommand) Execute(ctx command.CommandContext) {
	u := setup.FromContext(ctx)

	if u.InSetup() {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", translations.MessageAlreadyInSetup, utils.DEFAULT_PREFIX)
	} else {
		ctx.ReactWithCheck()

		u.Next()
		state := u.GetState()
		stage := state.GetStage()
		if stage != nil {
			// Psuedo-premium
			// TODO: TRANSLATE SETUP PROMPTS
			utils.SendEmbed(ctx.Worker, ctx.ChannelId, ctx.GuildId, utils.Green, "Setup", (*stage).Prompt(), nil, 120, true)
		}
	}
}
