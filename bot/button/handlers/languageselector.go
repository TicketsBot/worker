package handlers

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/i18n"
	"strings"
)

type LanguageSelectorHandler struct{}

func (h *LanguageSelectorHandler) Matcher() matcher.Matcher {
	return matcher.NewFuncMatcher(func(customId string) bool {
		return strings.HasPrefix(customId, "language-selector-")
	})
}

func (h *LanguageSelectorHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags: registry.SumFlags(registry.GuildAllowed),
	}
}

func (h *LanguageSelectorHandler) Execute(ctx *context.SelectMenuContext) {
	permissionLevel, err := ctx.UserPermissionLevel()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if permissionLevel < permission.Admin {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageNoPermission)
		return
	}

	if len(ctx.InteractionData.Values) == 0 {
		return
	}

	newLanguage := i18n.Language(ctx.InteractionData.Values[0])

	var valid bool
	for _, language := range i18n.LanguagesAlphabetical {
		if newLanguage == language {
			valid = true
			break
		}
	}

	// Infallible
	if !valid {
		ctx.ReplyRaw(customisation.Red, "Error", "Invalid language")
		return
	}

	if err := dbclient.Client.ActiveLanguage.Set(ctx.GuildId(), newLanguage.String()); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Reply(customisation.Green, i18n.TitleLanguage, i18n.MessageLanguageSuccess, i18n.FullNames[newLanguage], i18n.Flags[newLanguage])
}
