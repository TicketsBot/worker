package listeners

import (
	"context"
	"fmt"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/impl"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"github.com/rxdn/gdl/objects/interaction"
	"golang.org/x/sync/errgroup"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	channelPattern = regexp.MustCompile(`<#(\d+)>`)
	userPattern    = regexp.MustCompile(`<@!?(\d+)>`)
	rolePattern    = regexp.MustCompile(`<@&(\d+)>`)
)

func OnCommand(worker *worker.Context, e *events.MessageCreate) {
	if e.Author.Bot {
		return
	}

	// Ignore commands in DMs
	if e.GuildId == 0 {
		return
	}

	e.Member.User = e.Author

	var usedPrefix string

	if strings.HasPrefix(strings.ToLower(e.Content), utils.DEFAULT_PREFIX) {
		usedPrefix = utils.DEFAULT_PREFIX
	} else {
		// No need to query the custom prefix if we just the default prefix
		customPrefix, err := dbclient.Client.Prefix.Get(e.GuildId)
		if err != nil {
			sentry.Error(err)
			return
		}

		if customPrefix != "" && strings.HasPrefix(e.Content, customPrefix) {
			usedPrefix = customPrefix
		} else { // Not a command
			return
		}
	}

	split := strings.Split(e.Content, " ")
	root := split[0][len(usedPrefix):]

	args := make([]string, 0)
	if len(split) > 1 {
		for _, arg := range split[1:] {
			if arg != "" {
				args = append(args, arg)
			}
		}
	}

	var c command.Command
	for _, cmd := range impl.Commands {
		if cmd.Properties().InteractionOnly {
			continue
		}

		if strings.ToLower(cmd.Properties().Name) == strings.ToLower(root) || contains(cmd.Properties().Aliases, strings.ToLower(root)) {
			parent := cmd
			index := 0

			for {
				if len(args) > index {
					childName := args[index]
					found := false

					for _, child := range parent.Properties().Children {
						if child.Properties().InteractionOnly {
							continue
						}

						if strings.ToLower(child.Properties().Name) == strings.ToLower(childName) || contains(child.Properties().Aliases, strings.ToLower(childName)) {
							parent = child
							found = true
							index++
						}
					}

					if !found {
						break
					}
				} else {
					break
				}
			}

			var childArgs []string
			if len(args) > 0 {
				childArgs = args[index:]
			}

			args = childArgs
			c = parent
		}
	}

	if c == nil {
		return
	}

	properties := c.Properties()
	if properties.MainBotOnly && worker.IsWhitelabel {
		return
	}

	var blacklisted bool
	var premiumTier premium.PremiumTier
	var userPermissionLevel permcache.PermissionLevel

	group, _ := errgroup.WithContext(context.Background())

	// get blacklisted
	group.Go(func() (err error) {
		blacklisted, err = dbclient.Client.Blacklist.IsBlacklisted(e.GuildId, e.Author.Id)
		return
	})

	// get premium tier
	group.Go(func() error {
		premiumTier = utils.PremiumClient.GetTierByGuildId(e.GuildId, true, worker.Token, worker.RateLimiter)
		return nil
	})

	// get permission level
	group.Go(func() (err error) {
		userPermissionLevel, err = permcache.GetPermissionLevel(utils.ToRetriever(worker), e.Member, e.GuildId)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		// DEBUG START
		guild, err := worker.GetGuild(e.GuildId)
		if err != nil {
			return err
		}

		extra := map[string]interface{}{
			"member_id": strconv.FormatUint(e.Member.User.Id, 10),
			"level":    userPermissionLevel,
			"owner_id": strconv.FormatUint(guild.OwnerId, 10),
		}

		tags := map[string]string{
			"user_id":  strconv.FormatUint(e.Author.Id, 10),
			"guild_id": strconv.FormatUint(e.GuildId, 10),
		}

		sentry.LogWithTags("permission log", extra, tags)
		// DEBUG END
		return
	})

	if err := group.Wait(); err != nil {
		sentry.Error(err)
		return
	}

	// Ensure user isn't blacklisted
	if blacklisted {
		utils.ReactWithCross(worker, e.ChannelId, e.Id)
		return
	}

	ctx := command.NewMessageContext(worker, e.Message, args, premiumTier, userPermissionLevel)

	if properties.PermissionLevel > userPermissionLevel {
		ctx.Reject()
		ctx.Reply(utils.Red, "Error", translations.MessageNoPermission)
		return
	}

	if properties.AdminOnly && !utils.IsBotAdmin(e.Author.Id) {
		ctx.Reject()
		ctx.Reply(utils.Red, "Error", translations.MessageOwnerOnly)
		return
	}

	if properties.HelperOnly && !utils.IsBotHelper(e.Author.Id) {
		ctx.Reject()
		ctx.Reply(utils.Red, "Error", translations.MessageNoPermission)
		return
	}

	if properties.PremiumOnly && premiumTier == premium.None {
		ctx.Reject()
		ctx.Reply(utils.Red, "Premium Only Command", translations.MessagePremium)
		return
	}

	// parse args
	parsedArguments := make([]interface{}, len(properties.Arguments))

	var argsIndex int
	for i, argument := range properties.Arguments {
		if !argument.MessageCompatible {
			parsedArguments[i] = nil
		}

		if argsIndex >= len(args) {
			if argument.Required {
				ctx.Reply(utils.Red, "Error", argument.InvalidMessage)
				return
			}

			continue
		}

		// TODO: translate messages
		switch argument.Type {
		case interaction.OptionTypeString:
			parsedArguments[i] = strings.Join(args[argsIndex:], " ")
			argsIndex = len(args)
		case interaction.OptionTypeInteger:
			//goland:noinspection GoNilness
			raw := args[argsIndex]
			value, err := strconv.Atoi(raw)
			if err != nil {
				if argument.Required {
					ctx.Reply(utils.Red, "Error", argument.InvalidMessage)
					return
				} else {
					parsedArguments[i] = (*int)(nil)
					continue
				}
			}

			parsedArguments[i] = value
			argsIndex++
		case interaction.OptionTypeBoolean:
			//goland:noinspection GoNilness
			raw := args[argsIndex]
			value, err := strconv.ParseBool(raw)
			if err != nil {
				if argument.Required {
					ctx.Reply(utils.Red, "Error", argument.InvalidMessage)
					return
				} else {
					parsedArguments[i] = (*bool)(nil)
					continue
				}
			}

			parsedArguments[i] = value
			argsIndex++
		case interaction.OptionTypeUser:
			//goland:noinspection GoNilness
			match := userPattern.FindStringSubmatch(args[argsIndex])
			if len(match) < 2 {
				if argument.Required {
					ctx.Reply(utils.Red, "Error", argument.InvalidMessage)
					return
				} else {
					parsedArguments[i] = nil
					continue
				}
			}

			if userId, err := strconv.ParseUint(match[1], 10, 64); err == nil {
				parsedArguments[i] = userId
				argsIndex++
			} else {
				if argument.Required {
					ctx.Reply(utils.Red, "Error", argument.InvalidMessage)
					return
				} else {
					parsedArguments[i] = nil
					continue
				}
			}
		case interaction.OptionTypeChannel:
			//goland:noinspection GoNilness
			match := channelPattern.FindStringSubmatch(args[argsIndex])
			if len(match) < 2 {
				if argument.Required {
					ctx.Reply(utils.Red, "Error", argument.InvalidMessage)
					return
				} else {
					parsedArguments[i] = nil
					continue
				}
			}

			if channelId, err := strconv.ParseUint(match[1], 10, 64); err == nil {
				parsedArguments[i] = channelId
				argsIndex++
			} else {
				if argument.Required {
					ctx.Reply(utils.Red, "Error", argument.InvalidMessage)
					return
				} else {
					parsedArguments[i] = nil
					continue
				}
			}
		case interaction.OptionTypeRole:
			//goland:noinspection GoNilness
			match := rolePattern.FindStringSubmatch(args[argsIndex])
			if len(match) < 2 {
				if argument.Required {
					ctx.Reply(utils.Red, "Error", argument.InvalidMessage)
					return
				} else {
					parsedArguments[i] = nil
					continue
				}
			}

			if roleId, err := strconv.ParseUint(match[1], 10, 64); err == nil {
				parsedArguments[i] = roleId
				argsIndex++
			} else {
				if argument.Required {
					ctx.Reply(utils.Red, "Error", argument.InvalidMessage)
					return
				} else {
					parsedArguments[i] = nil
					continue
				}
			}
		}
	}

	e.Member.User = e.Author

	valueArgs := make([]reflect.Value, len(parsedArguments)+1)
	valueArgs[0] = reflect.ValueOf(&ctx)

	fn := reflect.TypeOf(c.GetExecutor())
	for i, arg := range parsedArguments {
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

	go reflect.ValueOf(c.GetExecutor()).Call(valueArgs)
	go statsd.Client.IncrementKey(statsd.KeyCommands)

	utils.DeleteAfter(worker, e.ChannelId, e.Id, utils.DeleteAfterSeconds)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
