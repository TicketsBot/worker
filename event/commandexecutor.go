package event

import (
	"errors"
	"fmt"
	"github.com/TicketsBot/common/premium"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
	"reflect"
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
					raw, ok := option.Value.(string)
					if !ok {
						return false, fmt.Errorf("option %s of type %d was not a string", option.Name, argument.Type)
					}

					id, err := strconv.ParseUint(raw, 10, 64)
					if err != nil {
						return false, err
					}

					args = append(args, id)
				}
			}
		}

		if !found {
			args = append(args, nil)
		}

		if !found && argument.Required {
			return false, fmt.Errorf("argument %s was missing for command %s", argument.Name, cmd.Properties().Name)
		}
	}

	properties := cmd.Properties()

	go func() {
		// get premium tier
		// TODO: guild id null check
		premiumLevel := utils.PremiumClient.GetTierByGuildId(data.GuildId.Value, true, ctx.Token, ctx.RateLimiter)

		interactionContext := command.NewSlashCommandContext(ctx, data, premiumLevel, responseCh)

		permLevel, err := interactionContext.UserPermissionLevel()
		if err != nil {
			interactionContext.HandleError(err)
			return
		}

		if properties.PermissionLevel > permLevel {
			interactionContext.Reject()
			interactionContext.Reply(utils.Red, "Error", translations.MessageNoPermission)
			return
		}

		if properties.AdminOnly && !utils.IsBotAdmin(interactionContext.UserId()) {
			interactionContext.Reject()
			interactionContext.Reply(utils.Red, "Error", translations.MessageOwnerOnly)
			return
		}

		if properties.HelperOnly && !utils.IsBotHelper(interactionContext.UserId()) {
			interactionContext.Reject()
			interactionContext.Reply(utils.Red, "Error", translations.MessageNoPermission)
			return
		}

		if properties.PremiumOnly && premiumLevel == premium.None {
			interactionContext.Reject()
			interactionContext.Reply(utils.Red, "Premium Only Command", translations.MessagePremium)
			return
		}

		var userId uint64
		if data.Member != nil {
			userId = data.Member.User.Id
		} else if data.User != nil {
			userId = data.User.Id
		} else { // ??????????????????????????????????
			interactionContext.HandleError(errors.New("userId was nil"))
			return
		}

		blacklisted, err := dbclient.Client.Blacklist.IsBlacklisted(data.GuildId.Value, userId)
		if err != nil {
			interactionContext.HandleError(err)
			return
		}

		if blacklisted {
			interactionContext.Reject()
			interactionContext.Reply(utils.Red, "Blacklisted", translations.MessageBlacklisted)
			return
		}

		valueArgs := make([]reflect.Value, len(args)+1)
		valueArgs[0] = reflect.ValueOf(&interactionContext)

		fn := reflect.TypeOf(cmd.GetExecutor())
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
		go statsd.Client.IncrementKey(statsd.KeySlashCommands)
		go statsd.Client.IncrementKey(statsd.KeyCommands)

		reflect.ValueOf(cmd.GetExecutor()).Call(valueArgs)
	}()

	return properties.DefaultEphemeral, nil
}
