package impl

import (
	"github.com/TicketsBot/common/permission"
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
	ctx.SendEmbed(utils.Green, "Vote", "Click here to vote for 24 hours of free premium:\nhttps://vote.ticketsbot.net")
	ctx.ReactWithCheck()
}
