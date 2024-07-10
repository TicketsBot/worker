package general

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"time"
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
		Timeout:          time.Second * 3,
	}
}

func (c InviteCommand) GetExecutor() interface{} {
	return c.Execute
}

func (InviteCommand) Execute(ctx registry.CommandContext) {
	ctx.Reply(customisation.Green, i18n.TitleInvite, i18n.MessageInvite)
}
