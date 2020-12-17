package admin

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"strconv"
)

type AdminSetMessageCommand struct {
}

// TODO: This is interaction only, but we don't want to show admin cmds
func (AdminSetMessageCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "setmessage",
		Description:     translations.HelpAdminSetMessage,
		Aliases:         []string{"sm"},
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:       true,
		MessageOnly:     true,
		/*
		 * Multiple strings, do manual arg parsing
		 *
		Arguments: command.Arguments(
			command.NewRequiredArgument("language", "Language", interaction.OptionTypeString, translations.MessageInvalidArgument),
			command.NewRequiredArgument("id", "ID of the message to update", interaction.OptionTypeInteger, translations.MessageInvalidArgument),
			command.NewRequiredArgument("value", "New value for the message", interaction.OptionTypeString, translations.MessageInvalidArgument),
		),*/
	}
}

func (c AdminSetMessageCommand) GetExecutor() interface{} {
	return c.Execute
}

// t!admin sm lang id value
func (AdminSetMessageCommand) Execute(ctx command.CommandContext) {
	msgCtx := ctx.(*command.MessageContext)

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
