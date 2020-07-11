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

type RemoveCommand struct {
}

func (RemoveCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "remove",
		Description:     translations.HelpRemove,
		PermissionLevel: permcache.Everyone,
		Category:        command.Tickets,
	}
}

func (RemoveCommand) Execute(ctx command.CommandContext) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!remove @User`",
		Inline: false,
	}

	if len(ctx.Message.Mentions) == 0 {
		ctx.SendEmbedWithFields(utils.Red, "Error", translations.MessageRemoveNoMembers, utils.FieldsToSlice(usageEmbed))
		ctx.ReactWithCross()
		return
	}

	// Get ticket struct
	ticket, err := dbclient.Client.Tickets.GetByChannel(ctx.ChannelId)
	if err != nil {
		ctx.ReactWithCross()
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	// Verify that the current channel is a real ticket
	if ticket.UserId == 0 {
		ctx.SendEmbed(utils.Red, "Error", translations.MessageNotATicketChannel)
		ctx.ReactWithCross()
		return
	}

	// Verify that the user is allowed to modify the ticket
	if ctx.UserPermissionLevel == 0 && ticket.UserId != ctx.Author.Id {
		ctx.SendEmbed(utils.Red, "Error", translations.MessageRemoveNoPermission)
		ctx.ReactWithCross()
		return
	}

	// verify that the user isn't trying to remove staff
	if ctx.MentionsStaff() {
		ctx.SendEmbed(utils.Red, "Error", translations.MessageRemoveCannotRemoveStaff)
		ctx.ReactWithCross()
		return
	}

	for _, user := range ctx.Message.Mentions {
		// Remove user from ticket in DB
		if err := dbclient.Client.TicketMembers.Delete(ctx.GuildId, ticket.Id, user.Id); err != nil {
			ctx.ReactWithCross()
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			return
		}

		// Remove user from ticket
		if err := ctx.Worker.EditChannelPermissions(ctx.ChannelId, channel.PermissionOverwrite{
			Id:    user.Id,
			Type:  channel.PermissionTypeMember,
			Allow: 0,
			Deny:  permission.BuildPermissions(permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks),
		}); err != nil {
			ctx.ReactWithCross()
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			return
		}
	}

	ctx.ReactWithCheck()
}
