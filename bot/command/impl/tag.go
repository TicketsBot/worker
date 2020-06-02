package impl

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"strings"
)

type TagCommand struct {
}

func (TagCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "tag",
		Description:     "Sends a message snippet",
		Aliases:         []string{"canned", "cannedresponse", "cr", "tags", "tag", "snippet", "c"},
		PermissionLevel: permission.Support,
		Category:        command.Tags,
	}
}

func (TagCommand) Execute(ctx command.CommandContext) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!tag [TagID]`",
		Inline: false,
	}

	if len(ctx.Args) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You must provide the ID of the tag. For more help with tag, visit <https://ticketsbot.net/tags>.", usageEmbed)
		ctx.ReactWithCross()
		return
	}

	tagId := strings.ToLower(ctx.Args[0])

	content, err := dbclient.Client.Tag.Get(ctx.GuildId, tagId)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		ctx.ReactWithCross()
		return
	}

	if content == "" {
		ctx.SendEmbed(utils.Red, "Error", "Invalid tag. For more help with tags, visit <https://ticketsbot.net/tags>.", usageEmbed)
		ctx.ReactWithCross()
		return
	}

	ticket, err := dbclient.Client.Tickets.GetByChannel(ctx.ChannelId)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	if ticket.UserId != 0 {
		mention := fmt.Sprintf("<@%d>", ticket.UserId)
		content = strings.Replace(content, "%user%", mention, -1)
	}

	ctx.ReactWithCheck()
	if _, err := ctx.Worker.CreateMessage(ctx.ChannelId, content); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}
}
