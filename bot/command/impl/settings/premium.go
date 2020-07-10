package settings

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/gofrs/uuid"
)

type PremiumCommand struct {
}

func (PremiumCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "premium",
		Description:     translations.HelpPremium,
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
	}
}

func (PremiumCommand) Execute(ctx command.CommandContext) {
	if len(ctx.Args) == 0 {
		if ctx.PremiumTier > premium.None {
			expiry, err := dbclient.Client.PremiumGuilds.GetExpiry(ctx.GuildId)
			if err != nil {
				ctx.ReactWithCross()
				sentry.ErrorWithContext(err, ctx.ToErrorContext())
				return
			}

			ctx.SendEmbed(utils.Red, "Premium", translations.MessageAlreadyPremium, expiry.UTC().String())
		} else {
			ctx.SendEmbed(utils.Red, "Premium", translations.MessagePremium)
		}
	} else {
		key, err := uuid.FromString(ctx.Args[0])

		if err != nil {
			ctx.SendEmbed(utils.Red, "Premium", translations.MessageInvalidPremiumKey)
			ctx.ReactWithCross()
			return
		}

		length, err := dbclient.Client.PremiumKeys.Delete(key)
		if err != nil {
			ctx.ReactWithCross()
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			return
		}

		if length == 0 {
			ctx.SendEmbed(utils.Red, "Premium", translations.MessageInvalidPremiumKey)
			ctx.ReactWithCross()
			return
		}

		if err := dbclient.Client.UsedKeys.Set(key, ctx.GuildId, ctx.Author.Id); err != nil {
			ctx.ReactWithCross()
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			return
		}

		if err := dbclient.Client.PremiumGuilds.Add(ctx.GuildId, length); err != nil {
			ctx.ReactWithCross()
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			return
		}

		data := premium.CachedTier{
			Tier:       int(premium.Premium),
			FromVoting: false,
		}

		if err = utils.PremiumClient.SetCachedTier(ctx.GuildId, data); err == nil {
			ctx.ReactWithCheck()
		} else {
			ctx.HandleError(err)
		}
	}
}
