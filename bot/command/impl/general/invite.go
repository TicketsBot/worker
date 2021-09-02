package general

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
)

type InviteCommand struct {
}

func (InviteCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:             "invite",
		Description:      i18n.MessageHelpInvite,
		Type:             interaction.ApplicationCommandTypeChatInput,
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
	ctx.Reply(constants.Green, "Invite", i18n.MessageInvite)
}
