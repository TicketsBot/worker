package utils

import (
	"context"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
)

func CanClose(ctx context.Context, cmd registry.CommandContext, ticket database.Ticket) bool {
	// Make sure user can close;
	// Get user's permissions level
	permissionLevel, err := cmd.UserPermissionLevel(ctx)
	if err != nil {
		cmd.HandleError(err)
		return false
	}

	if permissionLevel == permission.Everyone {
		usersCanClose, err := dbclient.Client.UsersCanClose.Get(ctx, cmd.GuildId())
		if err != nil {
			cmd.HandleError(err)
		}

		// If they are a normal user, don't let them close if users_can_close=false, or if they are not the opener
		if !usersCanClose || cmd.UserId() != ticket.UserId {
			return false
		}
	}

	return true
}
