package tickets

import (
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/permission"
)

type AddCommand struct {
}

func (AddCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "add",
		Description:     i18n.HelpAdd,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permcache.Everyone,
		Category:        command.Tickets,
		Arguments: command.Arguments(
			command.NewRequiredArgument("user", "User to add to the ticket", interaction.OptionTypeUser, i18n.MessageAddNoMembers),
			command.NewRequiredArgument("channel", "Channel to add the user to", interaction.OptionTypeChannel, i18n.MessageAddNoChannel),
		),
	}
}

func (c AddCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AddCommand) Execute(ctx registry.CommandContext, userId, channelId uint64) {
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
		ctx.ReplyWithFields(constants.Red, i18n.Error, i18n.MessageAddChannelNotTicket, utils.FieldsToSlice(usageEmbed))
		ctx.Reject()
		return
	}

	permissionLevel, err := ctx.UserPermissionLevel()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Verify that the user is allowed to modify the ticket
	if permissionLevel == permcache.Everyone && ticket.UserId != ctx.UserId() {
		ctx.Reply(constants.Red, i18n.Error, i18n.MessageAddNoPermission)
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

	ctx.Reply(constants.Green, i18n.TitleAdd, i18n.MessageAddSuccess, userId, channelId)
}
