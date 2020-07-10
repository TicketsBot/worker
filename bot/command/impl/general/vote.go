package general

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/utils"
)

type VoteCommand struct {
}

func (VoteCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "vote",
		Description:     "Gives you a link to vote for free premium",
		PermissionLevel: permission.Everyone,
		Category:        command.General,
	}
}

func (VoteCommand) Execute(ctx command.CommandContext) {
	ctx.SendEmbed(utils.Green, "Vote", translations.MessageVote)
	ctx.ReactWithCheck()
}
