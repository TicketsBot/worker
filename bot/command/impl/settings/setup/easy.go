package setup

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
)

type EasySetupCommand struct{}

func (EasySetupCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "ez",
		Description:     translations.HelpSetup,
		Aliases:         []string{"easy"},
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
	}
}
