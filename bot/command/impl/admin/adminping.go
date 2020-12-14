package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/utils"
	"time"
)

type AdminPingCommand struct {
}

func (AdminPingCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "ping",
		Description:     database.HelpAdminPing,
		Aliases:         []string{"latency"},
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
	}
}

func (c AdminPingCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminPingCommand) Execute(ctx command.CommandContext) {
	messageContext, ok := ctx.(*command.MessageContext)

	if ok {
		latency := time.Now().Sub(messageContext.Timestamp)
		ctx.ReplyRaw(utils.Green, "Admin", fmt.Sprintf("REST latency: `%dms`", latency))
		ctx.Accept()
	} else { // TODO: Take interaction ID -> get timestamp
		ctx.ReplyRaw(utils.Red, "Error", "Latency is not available for interactions")
		ctx.Reject()
	}
}
