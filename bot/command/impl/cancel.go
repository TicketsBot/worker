package impl

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/setup"
)

type CancelCommand struct {
}

func (CancelCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "cancel",
		Description:     "Cancels the setup process",
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
	}
}

func (CancelCommand) Execute(ctx command.CommandContext) {
	u := setup.FromContext(ctx)

	// Check if the user is in the setup process
	if !u.InSetup() {
		ctx.ReactWithCross()
		return
	}

	u.Cancel()
	ctx.ReactWithCheck()
}
