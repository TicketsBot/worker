package event

import (
	"context"
	"errors"
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	cmdcontext "github.com/TicketsBot/worker/bot/command/context"
	cmdregistry "github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/metrics/prometheus"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/config"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"golang.org/x/sync/errgroup"
	"runtime/debug"
)

// TODO: Command not found messages
// (defaultDefer, error)
func executeCommand(
	ctx *worker.Context,
	registry cmdregistry.Registry,
	data interaction.ApplicationCommandInteraction,
	responseCh chan interaction.ApplicationCommandCallbackData,
) (bool, error) {
	// data.Member is needed for permission level lookup
	if data.GuildId.Value == 0 || data.Member == nil {
		responseCh <- interaction.ApplicationCommandCallbackData{
			Content: "Commands in DMs are not currently supported. Please run this command in a server.",
		}
		return false, nil
	}

	cmd, ok := registry[data.Data.Name]
	if !ok {
		return false, fmt.Errorf("command %s does not exist", data.Data.Name)
	}

	options := data.Data.Options
	for len(options) > 0 && options[0].Value == nil { // Value and Options are mutually exclusive, value is never present on subcommands
		subCommand := options[0]

		var found bool
		for _, child := range cmd.Properties().Children {
			if child.Properties().Name == subCommand.Name {
				cmd = child
				found = true
				break
			}
		}

		if !found {
			return false, fmt.Errorf("subcommand %s does not exist for command %s", subCommand.Name, cmd.Properties().Name)
		}

		options = subCommand.Options
	}

	properties := cmd.Properties()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovering panicking goroutine while executing command %s: %v\n", properties.Name, r)
				debug.PrintStack()
			}
		}()

		// Parallelise queries
		group, _ := errgroup.WithContext(context.Background())

		// Get premium level
		var premiumLevel = premium.None
		group.Go(func() error {
			tier, err := utils.PremiumClient.GetTierByGuildId(data.GuildId.Value, true, ctx.Token, ctx.RateLimiter)
			if err != nil {
				// TODO: Better error handling
				// But do not hard fail, as Patreon / premium proxy may be experiencing some issues
				sentry.Error(err)
			} else {
				premiumLevel = tier
			}

			return nil
		})

		// Get permission level
		var permLevel = permission.Everyone
		group.Go(func() error {
			res, err := permission.GetPermissionLevel(utils.ToRetriever(ctx), *data.Member, data.GuildId.Value)
			if err != nil {
				return err
			}

			permLevel = res
			return nil
		})

		// Get guild blacklisted in guild
		var guildBlacklisted bool
		group.Go(func() error {
			res, err := dbclient.Client.ServerBlacklist.IsBlacklisted(data.GuildId.Value)
			if err != nil {
				return err
			}

			guildBlacklisted = res
			return nil
		})

		if err := group.Wait(); err != nil {
			errorId := sentry.Error(err)
			responseCh <- interaction.ApplicationCommandCallbackData{
				Content: fmt.Sprintf("An error occurred while processing this request (Error ID `%s`)", errorId),
			}
			return
		}

		if premiumLevel == premium.None && config.Conf.PremiumOnly {
			return
		}

		interactionContext := cmdcontext.NewSlashCommandContext(ctx, data, premiumLevel, responseCh)

		if guildBlacklisted {
			// TODO: Better message?
			interactionContext.Reply(customisation.Red, i18n.TitleBlacklisted, i18n.MessageBlacklisted)
			return
		}

		if properties.PermissionLevel > permLevel {
			interactionContext.Reply(customisation.Red, i18n.Error, i18n.MessageNoPermission)
			return
		}

		if properties.AdminOnly && !utils.IsBotAdmin(interactionContext.UserId()) {
			interactionContext.Reply(customisation.Red, i18n.Error, i18n.MessageOwnerOnly)
			return
		}

		if properties.HelperOnly && !utils.IsBotHelper(interactionContext.UserId()) {
			interactionContext.Reply(customisation.Red, i18n.Error, i18n.MessageNoPermission)
			return
		}

		if properties.PremiumOnly && premiumLevel == premium.None {
			interactionContext.Reply(customisation.Red, i18n.TitlePremiumOnly, i18n.MessagePremium)
			return
		}

		// Check for user blacklist - cannot parallelise as relies on permission level
		// If data.Member is nil, it does not matter, as it is not checked if the command is not executed in a guild
		blacklisted, err := interactionContext.IsBlacklisted()
		if err != nil {
			interactionContext.HandleError(err)
			return
		}

		if blacklisted {
			interactionContext.Reply(customisation.Red, i18n.TitleBlacklisted, i18n.MessageBlacklisted)
			return
		}

		statsd.Client.IncrementKey(statsd.KeySlashCommands)
		statsd.Client.IncrementKey(statsd.KeyCommands)
		prometheus.LogCommand(data.GuildId.Value, data.Data.Name)

		if err := callCommand(cmd, &interactionContext, options); err != nil {
			if errors.Is(err, ErrArgumentNotFound) {
				if ctx.IsWhitelabel {
					content := `This command registration is outdated. Please ask the server administrators to visit the whitelabel dashboard and press "Create Slash Commands" again.`
					embed := utils.BuildEmbedRaw(customisation.GetDefaultColour(customisation.Red), "Outdated Command", content, nil, premium.Whitelabel)
					res := command.NewEphemeralEmbedMessageResponse(embed)
					go func() { // Must be in a goroutine
						responseCh <- res.IntoApplicationCommandData()
					}()

					return
				} else {
					res := command.NewEphemeralTextMessageResponse("argument is missing")
					go func() {
						responseCh <- res.IntoApplicationCommandData()
					}()
				}
			} else {
				go interactionContext.HandleError(err)
				return
			}
		}
	}()

	return properties.DefaultEphemeral, nil
}
