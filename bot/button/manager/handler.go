package manager

import (
	"context"
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/button"
	"github.com/TicketsBot/worker/bot/button/registry"
	cmdcontext "github.com/TicketsBot/worker/bot/command/context"
	cmdregistry "github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/config"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/interaction/component"
	"golang.org/x/sync/errgroup"
	"time"
)

// Returns whether the handler may edit the message
func HandleInteraction(ctx context.Context, manager *ComponentInteractionManager, worker *worker.Context, data interaction.MessageComponentInteraction, responseCh chan button.Response) bool {
	// Safety checks - guild interactions only
	if data.GuildId.Value != 0 && data.Member == nil {
		return false
	}

	if data.GuildId.Value == 0 && data.User == nil {
		return false
	}

	lookupCtx, cancelLookupCtx := context.WithTimeout(ctx, time.Second*2)
	defer cancelLookupCtx()

	// Fetch premium tier
	var premiumTier = premium.None
	if data.GuildId.Value != 0 {
		// TODO: Re-architecture system to tie DMs to guilds
		tier, err := getPremiumTier(lookupCtx, worker, data.GuildId.Value)
		if err != nil {
			// TODO: Better handling
			// Allow executing to continue, assuming no premium
			sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{
				Guild:   data.GuildId.Value,
				Channel: data.ChannelId,
			})
		} else {
			premiumTier = tier
		}
	}

	if premiumTier == premium.None && config.Conf.PremiumOnly {
		return false
	}

	var cc cmdregistry.InteractionContext
	switch data.Data.Type() {
	case component.ComponentButton:
		cc = cmdcontext.NewButtonContext(ctx, worker, data, premiumTier, responseCh)
	case component.ComponentSelectMenu:
		cc = cmdcontext.NewSelectMenuContext(ctx, worker, data, premiumTier, responseCh)
	default:
		sentry.ErrorWithContext(fmt.Errorf("invalid message component type: %d", data.Data.ComponentType), errorcontext.WorkerErrorContext{
			Guild:   data.GuildId.Value,
			Channel: data.ChannelId,
		})
		return false
	}

	// Parallelise checks
	group, _ := errgroup.WithContext(lookupCtx)

	// Check if the user is blacklisted at guild / global level
	var userBlacklisted bool
	group.Go(func() (err error) {
		userBlacklisted, err = cc.IsBlacklisted(lookupCtx)
		return
	})

	// Check for guild-wide blacklist
	var guildBlacklisted = false
	if data.GuildId.Value != 0 {
		group.Go(func() (err error) {
			guildBlacklisted, err = dbclient.Client.ServerBlacklist.IsBlacklisted(lookupCtx, data.GuildId.Value)
			return
		})
	}

	if err := group.Wait(); err != nil {
		errorId := sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{
			Guild:   data.GuildId.Value,
			Channel: data.ChannelId,
		})

		cc.ReplyRaw(customisation.Red, "Error", fmt.Sprintf("An error occurred while processing this request (Error ID `%s`)", errorId))
		return false
	}

	if userBlacklisted || guildBlacklisted {
		cc.Reply(customisation.Red, i18n.TitleBlacklisted, i18n.MessageBlacklisted)
		return false
	}

	checkCtx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	switch data.Data.Type() {
	case component.ComponentButton:
		handler := manager.MatchButton(data.Data.AsButton().CustomId)
		if handler == nil {
			return false
		}

		shouldExecute, canEdit := doPropertiesChecks(checkCtx, data.GuildId.Value, cc, handler.Properties())
		if shouldExecute {
			go func() {
				defer close(responseCh)

				cc := cc.(*cmdcontext.ButtonContext)

				var cancel context.CancelFunc
				cc.Context, cancel = context.WithTimeout(cc.Context, handler.Properties().Timeout)
				defer cancel()

				handler.Execute(cc)
			}()
		}

		return canEdit
	case component.ComponentSelectMenu:
		handler := manager.MatchSelect(data.Data.AsSelectMenu().CustomId)
		if handler == nil {
			return false
		}

		shouldExecute, canEdit := doPropertiesChecks(checkCtx, data.GuildId.Value, cc, handler.Properties())
		if shouldExecute {
			go func() {
				defer close(responseCh)

				cc := cc.(*cmdcontext.SelectMenuContext)

				var cancel context.CancelFunc
				cc.Context, cancel = context.WithTimeout(cc.Context, handler.Properties().Timeout)
				defer cancel()

				handler.Execute(cc)
			}()
		}

		return canEdit
	default:
		sentry.ErrorWithContext(fmt.Errorf("invalid message component type: %d", data.Data.ComponentType), errorcontext.WorkerErrorContext{
			Guild:   data.GuildId.Value,
			Channel: data.ChannelId,
		})
		return false
	}
}

func getPremiumTier(ctx context.Context, worker *worker.Context, guildId uint64) (premium.PremiumTier, error) {
	// Psuedo premium if DM command
	if guildId == 0 {
		if worker.IsWhitelabel {
			return premium.Whitelabel, nil
		} else {
			return premium.Premium, nil
		}
	} else {
		premiumTier, err := utils.PremiumClient.GetTierByGuildId(ctx, guildId, true, worker.Token, worker.RateLimiter)
		if err != nil {
			return premium.None, err
		}

		return premiumTier, nil
	}
}

func doPropertiesChecks(ctx context.Context, guildId uint64, cmd cmdregistry.CommandContext, properties registry.Properties) (shouldExecute, canEdit bool) {
	if properties.PermissionLevel > permission.Everyone {
		permLevel, err := cmd.UserPermissionLevel(ctx)
		if err != nil {
			sentry.ErrorWithContext(err, cmd.ToErrorContext())
			return false, false
		}

		if permLevel < properties.PermissionLevel {
			cmd.Reply(customisation.Red, i18n.Error, i18n.MessageNoPermission)
			return false, false
		}
	}

	if guildId == 0 && !properties.HasFlag(registry.DMsAllowed) {
		cmd.Reply(customisation.Red, i18n.Error, i18n.MessageButtonGuildOnly)
		return false, false
	}

	if guildId != 0 && !properties.HasFlag(registry.GuildAllowed) {
		cmd.Reply(customisation.Red, i18n.Error, i18n.MessageButtonDMOnly)
		return false, false
	}

	return true, properties.HasFlag(registry.CanEdit)
}
