package tickets

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/rest"
)

type RenameCommand struct {
}

func (RenameCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "rename",
		Description:     translations.HelpRename,
		PermissionLevel: permission.Support,
		Category:        command.Tickets,
		Arguments: command.Arguments(
			command.NewRequiredArgument("name", "New name for the ticket", interaction.OptionTypeString, translations.MessageRenameMissingName),
		),
	}
}

func (c RenameCommand) GetExecutor() interface{} {
	return c.Execute
}

func (RenameCommand) Execute(ctx command.CommandContext, name string) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!rename [ticket-name]`",
		Inline: false,
	}

	ticket, err := dbclient.Client.Tickets.GetByChannel(ctx.ChannelId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Check this is a ticket channel
	if ticket.UserId == 0 {
		ctx.ReplyWithFields(utils.Red, "Rename", translations.MessageNotATicketChannel, utils.FieldsToSlice(usageEmbed))
		return
	}

	data := rest.ModifyChannelData{
		Name: name,
	}

	if _, err := ctx.Worker().ModifyChannel(ctx.ChannelId(), data); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Reply(utils.Green, "Rename", translations.MessageRenamed, ctx.ChannelId())
}
