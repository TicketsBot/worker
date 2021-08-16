package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/utils"
	"time"
)

type AdminPingCommand struct {
}

func (AdminPingCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "ping",
		Description:     i18n.HelpAdminPing,
		Aliases:         []string{"latency"},
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		HelperOnly:      true,
		MessageOnly:     true,
	}
}

func (c AdminPingCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminPingCommand) Execute(ctx registry.CommandContext) {
	messageContext, ok := ctx.(*context.MessageContext)

	if ok {
		latency := time.Now().Sub(messageContext.Timestamp)
		ctx.ReplyRaw(utils.Green, "Admin", fmt.Sprintf("REST latency: `%dms`", latency))
		ctx.Accept()
	} else { // TODO: Take interaction ID -> get timestamp
		ctx.ReplyRaw(utils.Red, "Error", "Latency is not available for interactions")
		ctx.Reject()
	}
}
