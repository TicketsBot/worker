package general

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
)

type VoteCommand struct {
}

func (VoteCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:             "vote",
		Description:      i18n.HelpVote,
		Type:             interaction.ApplicationCommandTypeChatInput,
		PermissionLevel:  permission.Everyone,
		Category:         command.General,
		DefaultEphemeral: true,
	}
}

func (c VoteCommand) GetExecutor() interface{} {
	return c.Execute
}

func (VoteCommand) Execute(ctx registry.CommandContext) {
	ctx.Reply(constants.Green, "Vote", i18n.MessageVote)
	ctx.Accept()
}
