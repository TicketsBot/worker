package tickets

import (
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel"
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
		),
	}
}

func (c AddCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AddCommand) Execute(ctx registry.CommandContext, userId uint64) {
	ticket, err := dbclient.Client.Tickets.GetByChannelAndGuild(ctx.ChannelId(), ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Test valid ticket channel
	if ticket.Id == 0 {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageNotATicketChannel)
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
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageAddNoPermission)
		ctx.Reject()
		return
	}

	// Add user to ticket in DB
	if err := dbclient.Client.TicketMembers.Add(ctx.GuildId(), ticket.Id, userId); err != nil {
		ctx.HandleError(err)
		return
	}

	// ticket.ChannelId cannot be nil, as we get by channel id
	data := channel.PermissionOverwrite{
		Id:    userId,
		Type:  channel.PermissionTypeMember,
		Allow: permission.BuildPermissions(logic.StandardPermissions[:]...),
	}

	if err := ctx.Worker().EditChannelPermissions(*ticket.ChannelId, data); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.ReplyPermanent(customisation.Green, i18n.TitleAdd, i18n.MessageAddSuccess, userId, *ticket.ChannelId)
}
