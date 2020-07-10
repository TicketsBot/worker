package admin

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"strconv"
	"strings"
)

type AdminSetMessageCommand struct {
}

func (AdminSetMessageCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "setmessage",
		Description:     translations.HelpAdminSetMessage,
		Aliases:         []string{"sm"},
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:      true,
	}
}

// t!admin sm lang id value
func (AdminSetMessageCommand) Execute(ctx command.CommandContext) {
	if len(ctx.Args) < 3 {
		ctx.SendEmbedRaw(utils.Red, "Error", "Invalid syntax. Use `t!admin setmessage lang id value`")
		return
	}

	language := ctx.Args[0]
	if len(language) > 8 {
		ctx.SendEmbedRaw(utils.Red, "Error", "Language ID is too long")
		return
	}

	id, err := strconv.Atoi(ctx.Args[1])
	if err != nil {
		ctx.SendEmbedRaw(utils.Red, "Error", "Invalid message ID")
		return
	}

	value := strings.Join(ctx.Args[2:], " ")

	if err := dbclient.Client.Translations.Set(language, translations.MessageId(id), value); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.ReactWithCheck()
}
