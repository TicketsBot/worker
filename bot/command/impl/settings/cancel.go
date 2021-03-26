package settings

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/setup"
)

type CancelCommand struct {
}

func (CancelCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "cancel",
		Description:     translations.HelpCancel,
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
	}
}

func (c CancelCommand) GetExecutor() interface{} {
	return c.Execute
}

func (CancelCommand) Execute(ctx registry.CommandContext) {
	u := setup.FromContext(ctx)

	// Check if the user is in the setup process
	if !u.InSetup() {
		ctx.Reject()
		return
	}

	u.Cancel()
	ctx.Accept()
}
