package tickets

import (
	permcache "github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/interaction"
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
		Arguments: command.Arguments(
			command.NewRequiredArgument("user", "User to remove from the current ticket", interaction.OptionTypeUser, translations.MessageRemoveAdminNoMembers),
		),
	}
}

func (c RemoveCommand) GetExecutor() interface{} {
	return c.Execute
}

func (RemoveCommand) Execute(ctx command.CommandContext, userId uint64) {
	/*usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!remove @User`",
		Inline: false,
	}*/

	// Get ticket struct
	ticket, err := dbclient.Client.Tickets.GetByChannel(ctx.ChannelId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Verify that the current channel is a real ticket
	if ticket.UserId == 0 {
		ctx.Reply(utils.Red, "Error", translations.MessageNotATicketChannel)
		ctx.Reject()
		return
	}

	selfPermissionLevel, err := ctx.UserPermissionLevel()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Verify that the user is allowed to modify the ticket
	if selfPermissionLevel == permcache.Everyone && ticket.UserId != ctx.UserId() {
		ctx.Reply(utils.Red, "Error", translations.MessageRemoveNoPermission)
		ctx.Reject()
		return
	}

	// verify that the user isn't trying to remove staff
	member, err := ctx.Worker().GetGuildMember(ctx.GuildId(), userId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	permissionLevel, err := permcache.GetPermissionLevel(utils.ToRetriever(ctx.Worker()), member, ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if permissionLevel >= permcache.Everyone {
		ctx.Reply(utils.Red, "Error", translations.MessageRemoveCannotRemoveStaff)
		ctx.Reject()
		return
	}

	// Remove user from ticket in DB
	if err := dbclient.Client.TicketMembers.Delete(ctx.GuildId(), ticket.Id, userId); err != nil {
		ctx.HandleError(err)
		return
	}

	// Remove user from ticket
	if err := ctx.Worker().EditChannelPermissions(ctx.ChannelId(), channel.PermissionOverwrite{
		Id:    userId,
		Type:  channel.PermissionTypeMember,
		Allow: 0,
		Deny:  permission.BuildPermissions(permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks),
	}); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Accept()
}
