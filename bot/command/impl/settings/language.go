package settings

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/interaction/component"
	"github.com/schollz/progressbar/v3"
	"io/ioutil"
	"math"
	"strings"
	"unicode"
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
	}
}

func (c LanguageCommand) GetExecutor() interface{} {
	return c.Execute
}

func (c *LanguageCommand) Execute(ctx registry.CommandContext) {
	var languageList string
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

		languageList += fmt.Sprintf("%s **%s** `%s`\n", flag, i18n.FullNamesEnglish[language], strings.TrimSpace(bar.String()))
	}

	languageList = strings.TrimSuffix(languageList, "\n")

	helpWanted := utils.EmbedField(ctx.GuildId(), "ℹ️ Help Wanted", i18n.MessageLanguageHelpWanted, true)
	e := utils.BuildEmbed(ctx, customisation.Green, i18n.TitleLanguage, i18n.MessageLanguageCommand, utils.ToSlice(helpWanted), languageList)
	res := command.NewEphemeralEmbedMessageResponseWithComponents(e, buildComponents(ctx))

	_, _ = ctx.ReplyWith(res)
}

func buildComponents(ctx registry.CommandContext) []component.Component {
	components := make([]component.Component, int(math.Ceil(float64(len(i18n.LanguagesAlphabetical))/25.0)))

	var menu component.SelectMenu
	var firstLanguage, lastLanguage i18n.Language
	var i int
	for j, language := range i18n.LanguagesAlphabetical {
		if j%25 == 0 {
			if j != 0 {
				startLetter := unicode.ToUpper(rune(i18n.LocalesInverse[firstLanguage][0]))
				endLetter := unicode.ToUpper(rune(i18n.LocalesInverse[lastLanguage][0]))

				menu.Placeholder = ctx.GetMessage(i18n.MessageLanguageSelect, startLetter, endLetter)
				components[i] = component.BuildActionRow(component.BuildSelectMenu(menu))
				i++
			}

			remainingLanguages := len(i18n.LanguagesAlphabetical) - (i * 25)
			menu = component.SelectMenu{
				CustomId: fmt.Sprintf("language-selector-%d", i),
				Options:  make([]component.SelectOption, utils.Min(remainingLanguages, 25)),
			}

			firstLanguage = language
		}

		menu.Options[j%25] = component.SelectOption{
			Label:       i18n.FullNamesEnglish[language],
			Description: i18n.FullNames[language],
			Value:       string(language),
			Emoji:       utils.BuildEmoji(i18n.Flags[language]),
			Default:     false,
		}

		lastLanguage = language
	}

	if len(menu.Options) > 0 {
		startLetter := unicode.ToUpper(rune(i18n.LocalesInverse[firstLanguage][0]))
		endLetter := unicode.ToUpper(rune(i18n.LocalesInverse[lastLanguage][0]))

		menu.Placeholder = ctx.GetMessage(i18n.MessageLanguageSelect, startLetter, endLetter)
		components[i] = component.BuildActionRow(component.BuildSelectMenu(menu))
	}

	return components
}
