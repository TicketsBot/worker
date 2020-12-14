package tickets

import (
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/permission"
)

type AddCommand struct {
}

func (AddCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "add",
		Description:     translations.HelpAdd,
		PermissionLevel: permcache.Everyone,
		Category:        command.Tickets,
		Arguments: command.Arguments(
			command.NewRequiredArgument("user", "User to add to the ticket", interaction.OptionTypeUser, translations.MessageAddNoMembers),
			command.NewRequiredArgument("channel", "Channel to add the user to", interaction.OptionTypeChannel, translations.MessageAddNoChannel),
		),
	}
}

func (c AddCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AddCommand) Execute(ctx command.CommandContext, userId, channelId uint64) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!add @User #ticket-channel`",
		Inline: false,
	}

	ticket, err := dbclient.Client.Tickets.GetByChannel(channelId)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	// 2 in 1: verify guild is the same & the channel is valid
	if ticket.GuildId != ctx.GuildId() {
		ctx.ReplyWithFields(utils.Red, "Error", translations.MessageAddChannelNotTicket, utils.FieldsToSlice(usageEmbed))
		ctx.Reject()
		return
	}

	// Get ticket ID
	owner := make(chan uint64)

	// Verify that the user is allowed to modify the ticket
	if ctx.UserPermissionLevel() == permcache.Everyone && <-owner != ctx.UserId() {
		ctx.Reply(utils.Red, "Error", translations.MessageAddNoPermission)
		ctx.Reject()
		return
	}

	// Add user to ticket in DB
	if err := dbclient.Client.TicketMembers.Add(ctx.GuildId(), ticket.Id, userId); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	if err := ctx.Worker().EditChannelPermissions(channelId, channel.PermissionOverwrite{
		Id:    userId,
		Type:  channel.PermissionTypeMember,
		Allow: permission.BuildPermissions(permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks),
	}); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	ctx.Accept()
}
