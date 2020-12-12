package tags

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
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
		Description:     translations.HelpTag,
		Aliases:         []string{"canned", "cannedresponse", "cr", "tags", "tag", "snippet", "c"},
		//Children:        []command.Command{ManageTagsListCommand{}, ManageTagsDeleteCommand{}, ManageTagsAddCommand{}},
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
		ctx.SendEmbedWithFields(utils.Red, "Error", translations.MessageTagInvalidArguments, utils.FieldsToSlice(usageEmbed))
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
		ctx.SendEmbedWithFields(utils.Red, "Error", translations.MessageTagInvalidTag, utils.FieldsToSlice(usageEmbed))
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

	_ = ctx.Worker.DeleteMessage(ctx.ChannelId, ctx.Id)

	if _, err := ctx.Worker.CreateMessage(ctx.ChannelId, content); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}
}
