package general

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/utils"
)

type VoteCommand struct {
}

func (VoteCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "vote",
		Description:     translations.HelpVote,
		PermissionLevel: permission.Everyone,
		Category:        command.General,
	}
}

func (c VoteCommand) GetExecutor() interface{} {
	return c.Execute
}

func (VoteCommand) Execute(ctx registry.CommandContext) {
	ctx.Reply(utils.Green, "Vote", translations.MessageVote)
	ctx.Accept()
}
