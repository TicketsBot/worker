package utils

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
)

func CanClose(ctx registry.CommandContext, ticket database.Ticket) bool  {
	// Make sure user can close;
	// Get user's permissions level
	permissionLevel, err := ctx.UserPermissionLevel()
	if err != nil {
		ctx.HandleError(err)
		return false
	}

	if permissionLevel == permission.Everyone {
		usersCanClose, err := dbclient.Client.UsersCanClose.Get(ctx.GuildId())
		if err != nil {
			ctx.HandleError(err)
		}

		// If they are a normal user, don't let them close if users_can_close=false, or if they are not the opener
		if !usersCanClose || ctx.UserId() != ticket.UserId {
			return false
		}
	}

	return true
}
