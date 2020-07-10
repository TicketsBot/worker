package tickets

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/rest"
	"strings"
)

type RenameCommand struct {
}

func (RenameCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "rename",
		Description:     "Renames the current ticket",
		PermissionLevel: permission.Support,
		Category:        command.Tickets,
	}
}

func (RenameCommand) Execute(ctx command.CommandContext) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!rename [ticket-name]`",
		Inline: false,
	}

	ticket, err := dbclient.Client.Tickets.GetByChannel(ctx.ChannelId)
	if err != nil {
		ctx.ReactWithCross()
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	// Check this is a ticket channel
	if ticket.UserId == 0 {
		ctx.SendEmbedWithFields(utils.Red, "Rename", translations.MessageNotATicketChannel, utils.FieldsToSlice(usageEmbed))
		return
	}

	if len(ctx.Args) == 0 {
		ctx.SendEmbedWithFields(utils.Red, "Rename", translations.MessageRenameMissingName, utils.FieldsToSlice(usageEmbed))
		return
	}

	name := strings.Join(ctx.Args, " ")
	data := rest.ModifyChannelData{
		Name: name,
	}

	if _, err := ctx.Worker.ModifyChannel(ctx.ChannelId, data); err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext()) // Probably 403
		return
	}

	ctx.SendEmbed(utils.Green, "Rename", translations.MessageRenamed, ctx.ChannelId)
}
