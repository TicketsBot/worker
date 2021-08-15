package admin

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/utils"
	"strconv"
)

type AdminSetMessageCommand struct {
}

// TODO: This is interaction only, but we don't want to show admin cmds
func (AdminSetMessageCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "setmessage",
		Description:     i18n.HelpAdminSetMessage,
		Aliases:         []string{"sm"},
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:       true,
		MessageOnly:     true,
		/*
		 * Multiple strings, do manual arg parsing
		 *
		Arguments: command.Arguments(
			command.NewRequiredArgument("language", "Language", interaction.OptionTypeString, i18n.MessageInvalidArgument),
			command.NewRequiredArgument("id", "ID of the message to update", interaction.OptionTypeInteger, i18n.MessageInvalidArgument),
			command.NewRequiredArgument("value", "New value for the message", interaction.OptionTypeString, i18n.MessageInvalidArgument),
		),*/
	}
}

func (c AdminSetMessageCommand) GetExecutor() interface{} {
	return c.Execute
}

// t!admin sm lang id value
func (AdminSetMessageCommand) Execute(ctx registry.CommandContext) {
	msgCtx := ctx.(*context.MessageContext)

	if len(msgCtx.Args) < 3 {
		ctx.ReplyRaw(utils.Red, "Error", "t!admin sm lang id value")
		return
	}

	language := msgCtx.Args[0]
	value := msgCtx.Args[2]

	id, err := strconv.Atoi(msgCtx.Args[1])
	if err != nil {
		ctx.ReplyRaw(utils.Red, "Error", "t!admin sm lang id value")
		return
	}

	if err := dbclient.Client.Translations.Set(language, translations.MessageId(id), value); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Accept()
}
