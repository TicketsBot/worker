package event

import (
	"context"
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	commandContext "github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/metrics/prometheus"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"golang.org/x/sync/errgroup"
	"reflect"
	"runtime/debug"
	"strconv"
)

// TODO: Command not found messages
// (defaultDefer, error)
func executeCommand(
	ctx *worker.Context,
	registry registry.Registry,
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

	var args []interface{}
	for _, argument := range cmd.Properties().Arguments {
		if !argument.SlashCommandCompatible {
			args = append(args, nil)
			continue
		}

		var found bool
		for _, option := range options {
			if option.Name == argument.Name {
				found = true

				// Discord does not validate types server side, so we must or risk panicking
				switch argument.Type {
				case interaction.OptionTypeString:
					if _, ok := option.Value.(string); !ok {
						return false, fmt.Errorf("option %s of type %d was not a string", option.Name, argument.Type)
					}

					args = append(args, option.Value)

				case interaction.OptionTypeInteger:
					raw, ok := option.Value.(float64)
					if !ok {
						return false, fmt.Errorf("option %s of type %d was not an integer", option.Name, argument.Type)
					}

					args = append(args, int(raw))

				case interaction.OptionTypeBoolean:
					if _, ok := option.Value.(bool); !ok {
						return false, fmt.Errorf("option %s of type %d was not a boolean", option.Name, argument.Type)
					}

					args = append(args, option.Value)

				// Parse snowflakes
				case interaction.OptionTypeUser:
					fallthrough
				case interaction.OptionTypeChannel:
					fallthrough
				case interaction.OptionTypeRole:
					fallthrough
				case interaction.OptionTypeMentionable:
					raw, ok := option.Value.(string)
					if !ok {
						return false, fmt.Errorf("option %s of type %d was not a string", option.Name, argument.Type)
					}

					id, err := strconv.ParseUint(raw, 10, 64)
					if err != nil {
						return false, err
					}

					args = append(args, id)
				case interaction.OptionTypeNumber:
					raw, ok := option.Value.(float64)
					if !ok {
						return false, fmt.Errorf("option %s of type %d was not an number", option.Name, argument.Type)
					}

					args = append(args, raw)
				default:
					return false, fmt.Errorf("unknown argument type: %d", argument.Type)
				}
			}
		}

		if !found {
			args = append(args, nil)
		}

		if !found && argument.Required {
			if ctx.IsWhitelabel {
				content := `This command registration is outdated. Please ask the server administrators to visit the whitelabel dashboard and press "Create Slash Commands" again.`
				embed := utils.BuildEmbedRaw(customisation.GetDefaultColour(customisation.Red), "Outdated Command", content, nil, premium.Whitelabel)
				res := command.NewEphemeralEmbedMessageResponse(embed)
				go func() { // Must be in a goroutine
					responseCh <- res.IntoApplicationCommandData()
				}()

				return false, nil
			} else {
				return false, fmt.Errorf("argument %s was missing for command %s", argument.Name, cmd.Properties().Name)
			}
		}
	}

	properties := cmd.Properties()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovering panicking goroutine while executing command %s: %v\n", properties.Name, r)
				debug.PrintStack()

				fmt.Printf("Command: %s\nArgs: %v\nData: %v\n", cmd.Properties().Name, args, data)
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

		interactionContext := commandContext.NewSlashCommandContext(ctx, data, premiumLevel, responseCh)

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

		fn := reflect.TypeOf(cmd.GetExecutor())
		if len(args) != fn.NumIn()-1 { // - 1 since command context is provided
			interactionContext.ReplyRaw(customisation.Red, "Error", "Argument count mismatch: Try creating slash commands again")
			return
		}

		valueArgs := make([]reflect.Value, len(args)+1)
		valueArgs[0] = reflect.ValueOf(&interactionContext)

		for i, arg := range args {
			var value reflect.Value
			if properties.Arguments[i].Required && arg != nil {
				value = reflect.ValueOf(arg)
			} else {
				if arg == nil {
					value = reflect.ValueOf(arg)
				} else {
					value = reflect.New(reflect.TypeOf(arg))
					tmp := value.Elem()
					tmp.Set(reflect.ValueOf(arg))
				}
			}

			if !value.IsValid() {
				value = reflect.New(fn.In(i + 1)).Elem()
			}

			valueArgs[i+1] = value
		}

		// Goroutine because recording metrics is blocking
		go func() {
			statsd.Client.IncrementKey(statsd.KeySlashCommands)
			statsd.Client.IncrementKey(statsd.KeyCommands)
			prometheus.LogCommand(data.GuildId.Value, data.Data.Name)
		}()

		reflect.ValueOf(cmd.GetExecutor()).Call(valueArgs)
	}()

	return properties.DefaultEphemeral, nil
}
