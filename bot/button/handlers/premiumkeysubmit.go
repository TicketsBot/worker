package handlers

import (
	"errors"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/gofrs/uuid"
)

type PremiumKeySubmitHandler struct{}

func (h *PremiumKeySubmitHandler) Matcher() matcher.Matcher {
	return matcher.NewSimpleMatcher("premium_key_modal")
}

func (h *PremiumKeySubmitHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags: registry.SumFlags(registry.GuildAllowed, registry.CanEdit),
	}
}

func (h *PremiumKeySubmitHandler) Execute(ctx *context.ModalContext) {
	permLevel, err := ctx.UserPermissionLevel()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if permLevel < permission.Admin {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageNoPermission)
		return
	}

	key, ok := ctx.GetInput("key")
	if !ok {
		ctx.HandleError(errors.New("key not found")) // Infallible providing non malicious
		return
	}

	parsed, err := uuid.FromString(key)
	if err != nil {
		ctx.EditWith(customisation.Red, i18n.TitlePremium, i18n.MessageInvalidPremiumKey)
		return
	}

	length, premiumTypeRaw, err := dbclient.Client.PremiumKeys.Delete(parsed)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if length == 0 {
		ctx.EditWith(customisation.Red, i18n.Error, i18n.MessageInvalidPremiumKey)
		return
	}

	premiumType := premium.PremiumTier(premiumTypeRaw)

	if err := dbclient.Client.UsedKeys.Set(parsed, ctx.GuildId(), ctx.UserId()); err != nil {
		ctx.HandleError(err)
		return
	}

	if premiumType == premium.Premium {
		if err := dbclient.Client.PremiumGuilds.Add(ctx.GuildId(), length); err != nil {
			ctx.HandleError(err)
			return
		}
	} else if premiumType == premium.Whitelabel { // TODO: Ensure user is admin
		if err := dbclient.Client.WhitelabelUsers.Add(ctx.UserId(), length); err != nil {
			ctx.HandleError(err)
			return
		}
	}

	data := premium.CachedTier{
		Tier:   int8(premiumTypeRaw),
		Source: premium.SourcePremiumKey,
	}

	if err = utils.PremiumClient.SetCachedTier(ctx.GuildId(), data); err == nil {
		ctx.EditWith(customisation.Green, i18n.TitlePremium, i18n.MessagePremiumSuccess, int(length.Hours()/24))
	} else {
		ctx.HandleError(err)
	}
}
