package settings

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"strings"
)

type LanguageCommand struct {
}

func (LanguageCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "language",
		Description:     translations.HelpLanguage,
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewRequiredArgument("language", "The country-code of the language to switch to", interaction.OptionTypeString),
		),
	}
}

func (c LanguageCommand) GetExecutor() interface{} {
	return c.Execute
}

func (c LanguageCommand) Execute(ctx command.CommandContext) {
	if len(ctx.Args) == 0 {
		c.sendInvalidMessage(ctx)
		return
	}

	newLanguage := ctx.Args[0]

	var valid bool
	for language, flag := range translations.Flags {
		if newLanguage == string(language) || newLanguage == flag {
			if err := dbclient.Client.ActiveLanguage.Set(ctx.GuildId, language); err != nil {
				ctx.HandleError(err)
			}

			valid = true
			break
		}
	}

	if !valid {
		c.sendInvalidMessage(ctx)
		return
	}

	ctx.ReactWithCheck()
}

func (LanguageCommand) sendInvalidMessage(ctx command.CommandContext) {
	example := embed.EmbedField{
		Name:   "Example",
		Value:  fmt.Sprintf("`%slanguage en`\n`%slanguage fr`\n`%slanguage de`", utils.DEFAULT_PREFIX, utils.DEFAULT_PREFIX, utils.DEFAULT_PREFIX),
		Inline: false,
	}

	var list string
	for language, flag := range translations.Flags {
		list += fmt.Sprintf("%s `%s\n`", flag, language)
	}
	list = strings.TrimSuffix(list, "\n")

	ctx.SendEmbedWithFields(utils.Red, "Error", translations.MessageLanguageInvalidLanguage, utils.FieldsToSlice(example), list)
	ctx.ReactWithCross()
}
