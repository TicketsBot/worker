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
	for _, locale := range i18n.Locales {
		if locale.Coverage == 0 {
			continue
		}

		bar := progressbar.NewOptions(100,
			progressbar.OptionSetWriter(ioutil.Discard),
			progressbar.OptionSetWidth(12),
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
		_ = bar.Set(locale.Coverage)

		languageList += fmt.Sprintf("%s **%s** `%s`\n", locale.FlagEmoji, locale.EnglishName, strings.TrimSpace(bar.String()))
	}

	languageList = strings.TrimSuffix(languageList, "\n")

	helpWanted := utils.EmbedField(ctx.GuildId(), "ℹ️ Help Wanted", i18n.MessageLanguageHelpWanted, true)
	e := utils.BuildEmbed(ctx, customisation.Green, i18n.TitleLanguage, i18n.MessageLanguageCommand, utils.ToSlice(helpWanted), languageList)
	res := command.NewEphemeralEmbedMessageResponseWithComponents(e, buildComponents(ctx))

	_, _ = ctx.ReplyWith(res)
}

func buildComponents(ctx registry.CommandContext) []component.Component {
	components := make([]component.Component, int(math.Ceil(float64(len(i18n.Locales))/25.0)))

	var menu component.SelectMenu
	var firstLocale, lastLocale *i18n.Locale
	var i int
	for j, locale := range i18n.Locales {
		if len(locale.Messages) == 0 {
			continue
		}

		if j%25 == 0 {
			if j != 0 {
				var startLetter, endLetter rune
				if firstLocale != nil { // should never be nil, but just in case
					startLetter = unicode.ToUpper(rune(firstLocale.IsoLongCode[0]))
				}

				if lastLocale != nil { // should never be nil, but just in case
					endLetter = unicode.ToUpper(rune(lastLocale.IsoLongCode[0]))
				}

				menu.Placeholder = ctx.GetMessage(i18n.MessageLanguageSelect, startLetter, endLetter)
				components[i] = component.BuildActionRow(component.BuildSelectMenu(menu))
				i++
			}

			remainingLanguages := len(i18n.Locales) - (i * 25)
			menu = component.SelectMenu{
				CustomId: fmt.Sprintf("language-selector-%d", i),
				Options:  make([]component.SelectOption, utils.Min(remainingLanguages, 25)),
			}

			firstLocale = locale
		}

		menu.Options[j%25] = component.SelectOption{
			Label:       locale.EnglishName,
			Description: locale.LocalName,
			Value:       locale.IsoShortCode,
			Emoji:       utils.BuildEmoji(locale.FlagEmoji),
			Default:     false,
		}

		lastLocale = locale
	}

	if len(menu.Options) > 0 {
		var startLetter, endLetter rune
		if firstLocale != nil { // should never be nil, but just in case
			startLetter = unicode.ToUpper(rune(firstLocale.IsoLongCode[0]))
		}

		if lastLocale != nil { // should never be nil, but just in case
			endLetter = unicode.ToUpper(rune(lastLocale.IsoLongCode[0]))
		}

		menu.Placeholder = ctx.GetMessage(i18n.MessageLanguageSelect, startLetter, endLetter)
		components[i] = component.BuildActionRow(component.BuildSelectMenu(menu))
	}

	return components
}
