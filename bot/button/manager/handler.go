package manager

import (
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
)

// Returns whether the handler may edit the message
func HandleInteraction(manager *ButtonManager, worker *worker.Context, data interaction.ButtonInteraction, editCh chan command.MessageResponse) bool {
	// Safety checks
	if data.GuildId.Value != 0 && data.Member == nil {
		return false
	}

	if data.GuildId.Value == 0 && data.User == nil {
		return false
	}

	handler := manager.Match(data.Data.CustomId)
	if handler == nil {
		return false
	}

	premiumTier, err := getPremiumTier(worker, data)
	if err != nil {
		sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{
			Guild:   data.GuildId.Value,
			Channel: data.ChannelId,
		})
		return false
	}

	ctx := context.NewButtonContext(worker, data, premiumTier, editCh)
	properties := handler.Properties()
	if data.GuildId.Value == 0 && !properties.HasFlag(registry.DMsAllowed) {
		ctx.Reply(constants.Red, i18n.Error, i18n.MessageButtonGuildOnly)
		return false
	}

	if data.GuildId.Value != 0 && !properties.HasFlag(registry.GuildAllowed) {
		ctx.Reply(constants.Red, i18n.Error, i18n.MessageButtonDMOnly)
		return false
	}

	go handler.Execute(ctx)

	return properties.HasFlag(registry.CanEdit)
}

func getPremiumTier(worker *worker.Context, data interaction.ButtonInteraction) (premium.PremiumTier, error) {
	// Psuedo premium if DM command
	if data.GuildId.Value == 0 {
		if worker.IsWhitelabel {
			return premium.Whitelabel, nil
		} else {
			return premium.Premium, nil
		}
	} else {
		premiumTier, err := utils.PremiumClient.GetTierByGuildId(data.GuildId.Value, true, worker.Token, worker.RateLimiter)
		if err != nil {
			return premium.None, err
		}

		return premiumTier, nil
	}
}