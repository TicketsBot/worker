package settings

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/gofrs/uuid"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/interaction"
	"time"
)

type PremiumCommand struct {
}

func (PremiumCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "premium",
		Description:     i18n.HelpPremium,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewOptionalArgument("key", "Premium key to activate", interaction.OptionTypeString, i18n.MessageInvalidPremiumKey),
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
				ctx.Reply(customisation.Red, i18n.TitlePremium, i18n.MessageAlreadyPremium, message.BuildTimestamp(expiry, message.TimestampStyleLongDateTime))
				return
			}
		}

		ctx.Reply(customisation.Red, i18n.TitlePremium, i18n.MessagePremium)
	} else {
		parsed, err := uuid.FromString(*key)

		if err != nil {
			ctx.Reply(customisation.Red, i18n.TitlePremium, i18n.MessageInvalidPremiumKey)
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
			ctx.Reply(customisation.Red, i18n.Error, i18n.MessageInvalidPremiumKey)
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
			Tier:   int8(premiumTypeRaw),
			Source: premium.SourcePremiumKey,
		}

		if err = utils.PremiumClient.SetCachedTier(ctx.GuildId(), data); err == nil {
			ctx.Reply(customisation.Green, i18n.TitlePremium, i18n.MessagePremiumSuccess, int(length.Hours()/24))
		} else {
			ctx.HandleError(err)
		}
	}
}
