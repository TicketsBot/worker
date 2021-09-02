package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"time"
)

type AdminPingCommand struct {
}

func (AdminPingCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "ping",
		Description:     i18n.HelpAdminPing,
		Type:            interaction.ApplicationCommandTypeChatInput,
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
		ctx.ReplyRaw(constants.Green, "Admin", fmt.Sprintf("REST latency: `%dms`", latency))
		ctx.Accept()
	} else { // TODO: Take interaction ID -> get timestamp
		ctx.ReplyRaw(constants.Red, "Error", "Latency is not available for interactions")
		ctx.Reject()
	}
}
