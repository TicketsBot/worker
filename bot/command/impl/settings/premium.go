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
	"github.com/rxdn/gdl/objects/interaction"
	"time"
)

type PremiumCommand struct {
}

func (PremiumCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "premium",
		Description:     translations.HelpPremium,
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewOptionalArgument("key", "Premium key to activate", interaction.OptionTypeString, translations.MessageInvalidPremiumKey),
		),
	}
}

func (c PremiumCommand) GetExecutor() interface{} {
	return c.Execute
}

func (PremiumCommand) Execute(ctx command.CommandContext, key *string) {
	if key == nil {
		if ctx.PremiumTier > premium.None {
			expiry, err := dbclient.Client.PremiumGuilds.GetExpiry(ctx.GuildId)
			if err != nil {
				ctx.ReactWithCross()
				sentry.ErrorWithContext(err, ctx.ToErrorContext())
				return
			}

			if expiry.After(time.Now()) {
				ctx.SendEmbed(utils.Red, "Premium", translations.MessageAlreadyPremium, expiry.UTC().String())
				return
			}
		}
		ctx.SendEmbed(utils.Red, "Premium", translations.MessagePremium)
	} else {
		parsed, err := uuid.FromString(*key)

		if err != nil {
			ctx.SendEmbed(utils.Red, "Premium", translations.MessageInvalidPremiumKey)
			ctx.ReactWithCross()
			return
		}

		length, err := dbclient.Client.PremiumKeys.Delete(parsed)
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

		if err := dbclient.Client.UsedKeys.Set(parsed, ctx.GuildId, ctx.Author.Id); err != nil {
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
