package manager

import (
	"context"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/button"
	cmdcontext "github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/config"
	"github.com/rxdn/gdl/objects/interaction"
	"time"
)

func HandleModalInteraction(ctx context.Context, manager *ComponentInteractionManager, worker *worker.Context, data interaction.ModalSubmitInteraction, responseCh chan button.Response) bool {
	// Safety checks
	if data.GuildId.Value != 0 && data.Member == nil {
		return false
	}

	if data.GuildId.Value == 0 && data.User == nil {
		return false
	}

	lookupCtx, cancelLookupCtx := context.WithTimeout(ctx, time.Second*2)
	defer cancelLookupCtx()

	premiumTier, err := getPremiumTier(lookupCtx, worker, data.GuildId.Value)
	if err != nil {
		sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{
			Guild:   data.GuildId.Value,
			Channel: data.ChannelId,
		})

		premiumTier = premium.None
	}

	if premiumTier == premium.None && config.Conf.PremiumOnly {
		return false
	}

	handler := manager.MatchModal(data.Data.CustomId)
	if handler == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(ctx, handler.Properties().Timeout)

	cc := cmdcontext.NewModalContext(ctx, worker, data, premiumTier, responseCh)
	shouldExecute, canEdit := doPropertiesChecks(lookupCtx, data.GuildId.Value, cc, handler.Properties())
	if shouldExecute {
		go func() {
			defer cancel()
			handler.Execute(cc)
		}()
	} else {
		cancel()
	}

	return canEdit
}
