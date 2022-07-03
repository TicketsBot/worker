package settings

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/schollz/progressbar/v3"
	"io/ioutil"
	"strings"
)

type LanguageCommand struct {
}

func (c LanguageCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "language",
		Description:     i18n.HelpLanguage,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewRequiredAutocompleteableArgument("language", "The country-code of the language to switch to", interaction.OptionTypeString, i18n.MessageLanguageInvalidLanguage, c.AutoCompleteHandler),
		),
	}
}

func (c LanguageCommand) GetExecutor() interface{} {
	return c.Execute
}

func (c LanguageCommand) Execute(ctx registry.CommandContext, newLanguage string) {
	var valid bool
	var newFlag string

	for language, flag := range i18n.Flags {
		if newLanguage == string(language) || newLanguage == flag {
			if err := dbclient.Client.ActiveLanguage.Set(ctx.GuildId(), language.String()); err != nil { // TODO: Don't wrap
				ctx.HandleError(err)
				return
			}

			newFlag = flag
			valid = true
			break
		}
	}

	if !valid {
		c.sendInvalidMessage(ctx)
		return
	}

	ctx.Reply(customisation.Green, i18n.TitleLanguage, i18n.MessageLanguageSuccess, newFlag)
}

func (LanguageCommand) sendInvalidMessage(ctx registry.CommandContext) {
	var list string
	for _, language := range i18n.LanguagesAlphabetical {
		coverage := i18n.GetCoverage(language)
		if coverage == 0 {
			continue
		}

		flag := i18n.Flags[language]

		bar := progressbar.NewOptions(100,
			progressbar.OptionSetWriter(ioutil.Discard),
			progressbar.OptionSetWidth(15),
			progressbar.OptionSetPredictTime(false),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "=",
				SaucerHead:    ">",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}),
		)
		_ = bar.Set(coverage)

		list += fmt.Sprintf("%s **%s** `%s`\n", flag, language, strings.TrimSpace(bar.String()))
	}

	list = strings.TrimSuffix(list, "\n")

	example := utils.EmbedFieldRaw("Example", fmt.Sprintf("`/language en`\n`/language fr`\n`/language de`"), true)
	helpWanted := utils.EmbedField(ctx.GuildId(), "ℹ️ Help Wanted", i18n.MessageLanguageHelpWanted, true)

	ctx.ReplyWithFields(customisation.Red, i18n.Error, i18n.MessageLanguageInvalidLanguage, utils.ToSlice(example, utils.BlankField(true), helpWanted), list)
	ctx.Accept()
}

func (c LanguageCommand) AutoCompleteHandler(data interaction.ApplicationCommandAutoCompleteInteraction, value string) (choices []interaction.ApplicationCommandOptionChoice) {
	valLower := strings.ToLower(value)

	// Don't iter map in order to keep alphabetical
	for _, code := range i18n.LanguagesAlphabetical {
		fullName := i18n.FullNames[code]
		if strings.HasPrefix(strings.ToLower(fullName), valLower) || strings.HasPrefix(strings.ToLower(code.String()), valLower) {
			choices = append(choices, interaction.ApplicationCommandOptionChoice{
				Name:  fullName,
				Value: code.String(),
			})
		}
	}

	if len(choices) > 25 {
		return choices[:25]
	} else {
		return choices
	}
}
