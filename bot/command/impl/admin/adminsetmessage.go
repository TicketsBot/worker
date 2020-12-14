package admin

import (
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/rxdn/gdl/objects/interaction"
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
		AdminOnly:       true,
		InteractionOnly: true,
		Arguments: command.Arguments(
			command.NewRequiredArgument("language", "Language", interaction.OptionTypeString, translations.MessageInvalidArgument),
			command.NewRequiredArgument("id", "ID of the message to update", interaction.OptionTypeInteger, translations.MessageInvalidArgument),
			command.NewRequiredArgument("value", "New value for the message", interaction.OptionTypeString, translations.MessageInvalidArgument),
		),
	}
}

func (c AdminSetMessageCommand) GetExecutor() interface{} {
	return c.Execute
}

// t!admin sm lang id value
func (AdminSetMessageCommand) Execute(ctx command.CommandContext, language string, id int, value string) {
	if err := dbclient.Client.Translations.Set(language, translations.MessageId(id), value); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Accept()
}
