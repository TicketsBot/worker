package impl

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/setup"
	"github.com/TicketsBot/worker/bot/utils"
)

type SetupCommand struct {
}

func (SetupCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "setup",
		Description:     "Allows you to easily configure the bot",
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
	}
}

func (SetupCommand) Execute(ctx command.CommandContext) {
	u := setup.FromContext(ctx)

	if u.InSetup() {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", fmt.Sprintf("You are already in setup mode (use `%scancel` to exit)", utils.DEFAULT_PREFIX))
	} else {
		ctx.ReactWithCheck()

		u.Next()
		state := u.GetState()
		stage := state.GetStage()
		if stage != nil {
			// Psuedo-premium
			utils.SendEmbed(ctx.Worker, ctx.ChannelId, utils.Green, "Setup", (*stage).Prompt(), nil, 120, true)
		}
	}
}
