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
	"time"
)

type LanguageSelectorHandler struct{}

func (h *LanguageSelectorHandler) Matcher() matcher.Matcher {
	return matcher.NewFuncMatcher(func(customId string) bool {
		return strings.HasPrefix(customId, "language-selector-")
	})
}

func (h *LanguageSelectorHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags:   registry.SumFlags(registry.GuildAllowed),
		Timeout: time.Second * 3,
	}
}

func (h *LanguageSelectorHandler) Execute(ctx *context.SelectMenuContext) {
	permissionLevel, err := ctx.UserPermissionLevel(ctx)
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

	newLocale, ok := i18n.MappedByIsoShortCode[ctx.InteractionData.Values[0]]
	// Infallible
	if !ok {
		ctx.ReplyRaw(customisation.Red, "Error", "Invalid language")
		return
	}

	if err := dbclient.Client.ActiveLanguage.Set(ctx, ctx.GuildId(), newLocale.IsoShortCode); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Reply(customisation.Green, i18n.TitleLanguage, i18n.MessageLanguageSuccess, newLocale.LocalName, newLocale.FlagEmoji)
}
