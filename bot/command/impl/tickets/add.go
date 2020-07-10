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
	"github.com/rxdn/gdl/permission"
)

type AddCommand struct {
}

func (AddCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "add",
		Description:     "Adds a user to a ticket",
		PermissionLevel: permcache.Everyone,
		Category:        command.Tickets,
	}
}

func (a AddCommand) Execute(ctx command.CommandContext) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!add @User #ticket-channel`",
		Inline: false,
	}

	// Check users are mentioned
	if len(ctx.Message.Mentions) == 0 {
		ctx.SendEmbed(utils.Red, "Error", translations.MessageAddNoMembers, usageEmbed)
		ctx.ReactWithCross()
		return
	}

	// Check channel is mentioned
	ticketChannel := ctx.GetChannelFromArgs()
	if ticketChannel == 0 {
		ctx.SendEmbed(utils.Red, "Error", translations.MessageAddNoChannel, usageEmbed)
		ctx.ReactWithCross()
		return
	}

	ticket, err := dbclient.Client.Tickets.GetByChannel(ticketChannel)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	// 2 in 1: verify guild is the same & the channel is valid
	if ticket.GuildId != ctx.GuildId {
		ctx.SendEmbed(utils.Red, "Error", translations.MessageAddChannelNotTicket, usageEmbed)
		ctx.ReactWithCross()
		return
	}

	// Get ticket ID
	owner := make(chan uint64)

	// Verify that the user is allowed to modify the ticket
	if ctx.UserPermissionLevel == 0 && <-owner != ctx.Author.Id {
		ctx.SendEmbed(utils.Red, "Error", translations.MessageAddNoPermission)
		ctx.ReactWithCross()
		return
	}

	for _, user := range ctx.Message.Mentions {
		// Add user to ticket in DB
		go func() {
			if err := dbclient.Client.TicketMembers.Add(ctx.GuildId, ticket.Id, user.Id); err != nil {
				sentry.ErrorWithContext(err, ctx.ToErrorContext())
			}
		}()

		if err := ctx.Worker.EditChannelPermissions(ticketChannel, channel.PermissionOverwrite{
			Id:    user.Id,
			Type:  channel.PermissionTypeMember,
			Allow: permission.BuildPermissions(permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks),
		}); err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		}
	}

	ctx.ReactWithCheck()
}
