package general

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/utils"
)

type InviteCommand struct {
}

func (InviteCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:             "invite",
		Description:      translations.MessageHelpInvite,
		PermissionLevel:  permission.Everyone,
		Category:         command.General,
		MainBotOnly:      true,
		DefaultEphemeral: true,
	}
}

func (c InviteCommand) GetExecutor() interface{} {
	return c.Execute
}

func (InviteCommand) Execute(ctx registry.CommandContext) {
	ctx.Reply(utils.Green, "Invite", translations.MessageInvite)
}
