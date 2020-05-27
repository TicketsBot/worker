package listeners

import (
	"context"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/impl"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/sentry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"golang.org/x/sync/errgroup"
	"strings"
)

func OnCommand(worker *worker.Context, e *events.MessageCreate) {
	if e.Author.Bot {
		return
	}

	// Ignore commands in DMs
	if e.GuildId == 0 {
		return
	}

	customPrefix, err := dbclient.Client.Prefix.Get(e.GuildId)
	if err != nil {
		sentry.Error(err)
	}

	var usedPrefix string

	if strings.HasPrefix(e.Content, utils.DEFAULT_PREFIX) {
		usedPrefix = utils.DEFAULT_PREFIX
	} else if strings.HasPrefix(e.Content, customPrefix) && customPrefix != "" {
		usedPrefix = customPrefix
	} else { // Not a command
		return
	}

	split := strings.Split(e.Content, " ")
	root := strings.TrimPrefix(split[0], usedPrefix)

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
		if strings.ToLower(cmd.Properties().Name) == strings.ToLower(root) || contains(cmd.Properties().Aliases, strings.ToLower(root)) {
			parent := cmd
			index := 0

			for {
				if len(args) > index {
					childName := args[index]
					found := false

					for _, child := range parent.Properties().Children {
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

	var blacklisted bool
	var premiumTier premium.PremiumTier

	group, _ := errgroup.WithContext(context.Background())

	// get blacklisted
	group.Go(func() (err error) {
		blacklisted, err = dbclient.Client.Blacklist.IsBlacklisted(e.GuildId, e.Author.Id)
		return
	})

	// get utils
	group.Go(func() error {
		premiumTier = utils.PremiumClient.GetTierByGuildId(e.GuildId, true, worker.Token, worker.RateLimiter)
		return nil
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

	e.Member.User = e.Author

	ctx := command.CommandContext{
		Worker:      worker,
		Message:     e.Message,
		Root:        root,
		Args:        args,
		PremiumTier: premiumTier,
		ShouldReact: true,
		IsFromPanel: false,
	}

	if c != nil {
		var permLevel permission.PermissionLevel
		{
			ch := make(chan permission.PermissionLevel)
			go ctx.GetPermissionLevel(ch)
			permLevel = <-ch
		}
		ctx.UserPermissionLevel = permLevel

		if c.Properties().PermissionLevel > permLevel {
			ctx.ReactWithCross()
			ctx.SendEmbed(utils.Red, "Error", utils.NO_PERMISSION)
			return
		}

		if c.Properties().AdminOnly && !utils.IsBotAdmin(e.Author.Id) {
			ctx.ReactWithCross()
			ctx.SendEmbed(utils.Red, "Error", "This command is reserved for the bot owner only")
			return
		}

		if c.Properties().HelperOnly && !utils.IsBotHelper(e.Author.Id) {
			ctx.ReactWithCross()
			ctx.SendEmbed(utils.Red, "Error", utils.NO_PERMISSION)
			return
		}

		if c.Properties().PremiumOnly && premiumTier == premium.None {
			ctx.ReactWithCross()
			ctx.SendEmbed(utils.Red, "PremiumTier Only Command", utils.PREMIUM_MESSAGE)
			return
		}

		go c.Execute(ctx)
		go statsd.IncrementKey(statsd.COMMANDS)

		utils.DeleteAfter(utils.SentMessage{Worker: worker, Message: &e.Message}, 30)
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
