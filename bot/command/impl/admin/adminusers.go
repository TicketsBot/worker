package admin

import (
	"context"
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
)

type AdminUsersCommand struct {
}

func (AdminUsersCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "users",
		Description:     i18n.HelpAdminUsers,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
		MessageOnly:     true,
	}
}

func (c AdminUsersCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminUsersCommand) Execute(ctx registry.CommandContext) {
	var count int

	query := `SELECT COUNT(DISTINCT "user_id") FROM members;`
	if err := ctx.Worker().Cache.QueryRow(context.Background(), query).Scan(&count); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.ReplyRaw(utils.Green, "Admin", fmt.Sprintf("Seen %d users", count))
}