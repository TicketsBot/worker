package admin

import (
	"context"
	"fmt"
	"github.com/TicketsBot/common/permission"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/utils"
)

type AdminUsersCommand struct {
}

func (AdminUsersCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "users",
		Description:     database.HelpAdminUsers,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
	}
}

func (AdminUsersCommand) Execute(ctx command.CommandContext) {
	var count int

	query := `SELECT COUNT(DISTINCT "user_id") FROM members;`
	if err := ctx.Worker.Cache.QueryRow(context.Background(), query).Scan(&count); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.SendEmbedRaw(utils.Green, "Admin", fmt.Sprintf("Seen %d users", count))
}