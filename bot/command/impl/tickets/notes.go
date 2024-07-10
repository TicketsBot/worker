package tickets

import (
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/rest"
	"strconv"
	"strings"
	"time"
)

type NotesCommand struct {
}

func (NotesCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:             "notes",
		Description:      i18n.HelpNotes,
		Type:             interaction.ApplicationCommandTypeChatInput,
		PermissionLevel:  permcache.Support,
		Category:         command.Tickets,
		DefaultEphemeral: true,
		Timeout:          time.Second * 7,
	}
}

func (c NotesCommand) GetExecutor() interface{} {
	return c.Execute
}

func (NotesCommand) Execute(ctx registry.CommandContext) {
	ticket, err := dbclient.Client.Tickets.GetByChannelAndGuild(ctx, ctx.ChannelId(), ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Test valid ticket channel
	if ticket.Id == 0 || ticket.ChannelId == nil {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageNotATicketChannel)
		return
	}

	if ticket.IsThread {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageNotesChannelModeOnly)
		return
	}

	var panel *database.Panel
	if ticket.PanelId != nil {
		tmp, err := dbclient.Client.Panel.GetById(ctx, *ticket.PanelId)
		if err != nil {
			ctx.HandleError(err)
			return
		}

		panel = &tmp
	}

	if ticket.NotesThreadId != nil {
		// Check if user is staff member
		// HasPermissionForTicket returns true if the user opened the ticket, but the command's properties enforces
		// requiring the user to be a staff member
		hasPermission, err := logic.HasPermissionForTicket(ctx, ctx.Worker(), ticket, ctx.UserId())
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if !hasPermission {
			ctx.Reply(customisation.Red, i18n.Error, i18n.MessageNoPermission)
			return
		}

		if err := ctx.Worker().AddThreadMember(*ticket.NotesThreadId, ctx.UserId()); err != nil {
			ctx.HandleError(err)
			return
		}

		ctx.Reply(customisation.Green, i18n.Success, i18n.MessageNotesAddedToExisting, *ticket.NotesThreadId)
	} else {
		allowedUsers, allowedRoles, err := logic.GetAllowedStaffUsersAndRoles(ctx, ctx.GuildId(), panel)
		if err != nil {
			ctx.HandleError(err)
			return
		}

		var b strings.Builder
		b.Grow(utils.Min(len(allowedRoles)*22+len(allowedUsers)*21, 2000)) // Provide size hint

		// Make sure user is added to the thread, as they may be the server owner but not have any of the staff roles
		b.WriteString("<@" + strconv.FormatUint(ctx.UserId(), 10) + ">")

		// Add roles first
		for _, roleId := range allowedRoles {
			mention := "<@&" + strconv.FormatUint(roleId, 10) + ">"

			if b.Len()+len(mention) > 2000 {
				break
			}

			_, _ = b.WriteString(mention) // Error is always nil
		}

		for _, roleId := range allowedUsers {
			mention := "<@" + strconv.FormatUint(roleId, 10) + ">"

			if b.Len()+len(mention) > 2000 {
				break
			}

			_, _ = b.WriteString(mention) // Error is always nil
		}

		editData := rest.EditMessageData{
			Content: b.String(),
		}

		thread, err := ctx.Worker().CreatePrivateThread(ctx.ChannelId(), ctx.GetMessage(i18n.MessageNotesThreadName), 10080, false)
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if err := dbclient.Client.Tickets.SetNotesThreadId(ctx, ticket.GuildId, ticket.Id, thread.Id); err != nil {
			ctx.HandleError(err)
			return
		}

		// Add staff to thread
		msg, err := ctx.Worker().CreateMessage(thread.Id, "Adding members...")
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if _, err := ctx.Worker().EditMessage(thread.Id, msg.Id, editData); err != nil {
			ctx.HandleError(err)
			return
		}

		if err := ctx.Worker().DeleteMessage(thread.Id, msg.Id); err != nil {
			ctx.HandleError(err)
			return
		}

		ctx.Reply(customisation.Green, i18n.Success, i18n.MessageNotesCreated, thread.Id)
	}
}
