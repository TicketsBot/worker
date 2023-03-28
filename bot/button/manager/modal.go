package manager

import (
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/button"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/rxdn/gdl/objects/interaction"
)

func HandleModalInteraction(manager *ComponentInteractionManager, worker *worker.Context, data interaction.ModalSubmitInteraction, responseCh chan button.Response) bool {
	// Safety checks
	if data.GuildId.Value != 0 && data.Member == nil {
		return false
	}

	if data.GuildId.Value == 0 && data.User == nil {
		return false
	}

	premiumTier, err := getPremiumTier(worker, data.GuildId.Value)
	if err != nil {
		sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{
			Guild:   data.GuildId.Value,
			Channel: data.ChannelId,
		})

		premiumTier = premium.None
	}

	handler := manager.MatchModal(data.Data.CustomId)
	if handler == nil {
		return false
	}

	ctx := context.NewModalContext(worker, data, premiumTier, responseCh)
	shouldExecute, canEdit := doPropertiesChecks(data.GuildId.Value, ctx, handler.Properties())
	if shouldExecute {
		go handler.Execute(ctx)
	}

	return canEdit
}

