package general

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/utils"
)

type InviteCommand struct {
}

func (InviteCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "invite",
		Description:     translations.MessageHelpInvite,
		PermissionLevel: permission.Everyone,
		Category:        command.General,
		MainBotOnly:     true,
	}
}

func (c InviteCommand) GetExecutor() interface{} {
	return c.Execute
}

func (InviteCommand) Execute(ctx command.CommandContext) {
	ctx.Reply(utils.Green, "Invite", translations.MessageInvite)
}
