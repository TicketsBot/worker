package settings

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/gofrs/uuid"
	"github.com/rxdn/gdl/objects/interaction"
	"time"
)

type PremiumCommand struct {
}

func (PremiumCommand) Properties() registry.Properties {
	return registry.Properties{
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

func (PremiumCommand) Execute(ctx registry.CommandContext, key *string) {
	if key == nil {
		if ctx.PremiumTier() > premium.None {
			expiry, err := dbclient.Client.PremiumGuilds.GetExpiry(ctx.GuildId())
			if err != nil {
				ctx.Reject()
				sentry.ErrorWithContext(err, ctx.ToErrorContext())
				return
			}

			if expiry.After(time.Now()) {
				ctx.Reply(utils.Red, "Premium", translations.MessageAlreadyPremium, expiry.UTC().String())
				return
			}
		}
		ctx.Reply(utils.Red, "Premium", translations.MessagePremium)
	} else {
		parsed, err := uuid.FromString(*key)

		if err != nil {
			ctx.Reply(utils.Red, "Premium", translations.MessageInvalidPremiumKey)
			ctx.Reject()
			return
		}

		length, premiumTypeRaw, err := dbclient.Client.PremiumKeys.Delete(parsed)
		if err != nil {
			ctx.Reject()
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			return
		}

		if length == 0 {
			ctx.Reply(utils.Red, "Premium", translations.MessageInvalidPremiumKey)
			ctx.Reject()
			return
		}

		premiumType := premium.PremiumTier(premiumTypeRaw)

		if err := dbclient.Client.UsedKeys.Set(parsed, ctx.GuildId(), ctx.UserId()); err != nil {
			ctx.Reject()
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			return
		}

		if premiumType == premium.Premium {
			if err := dbclient.Client.PremiumGuilds.Add(ctx.GuildId(), length); err != nil {
				ctx.HandleError(err)
				ctx.Reject()
				return
			}
		} else if premiumType == premium.Whitelabel {
			if err := dbclient.Client.WhitelabelUsers.Add(ctx.UserId(), length); err != nil {
				ctx.HandleError(err)
				ctx.Reject()
				return
			}
		}

		data := premium.CachedTier{
			Tier:       premiumTypeRaw,
			FromVoting: false,
		}

		if err = utils.PremiumClient.SetCachedTier(ctx.GuildId(), data); err == nil {
			ctx.Accept()
		} else {
			ctx.HandleError(err)
		}
	}
}
