package settings

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"time"
)

type SyncCommand struct {
}

func (SyncCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "sync",
		Description:     translations.HelpSync,
		PermissionLevel: permission.Admin,
		Category:        command.Settings,
	}
}

func (c SyncCommand) GetExecutor() interface{} {
	return c.Execute
}

func (s SyncCommand) Execute(ctx command.CommandContext) {
	if !utils.IsBotHelper(ctx.UserId()) {
		if s.isInCooldown(ctx.GuildId()) {
			ctx.ReplyRaw(utils.Red, "Sync", "This command is currently in cooldown")
			return
		}

		s.addCooldown(ctx.GuildId())
	}

	// Process deleted tickets
	ctx.ReplyPlain("Scanning for deleted ticket channels...")
	ctx.ReplyPlain(fmt.Sprintf("Completed **%d** ticket state synchronisation(s)", processDeletedTickets(ctx)))

	// Check any panels still exist
	ctx.ReplyPlain("Scanning for deleted panels...")
	ctx.ReplyPlain(fmt.Sprintf("Completed **%d** panel state synchronisation(s)", processDeletedPanels(ctx)))

	ctx.ReplyPlain("Sync complete!")
}

const cooldown = time.Minute

func (s SyncCommand) isInCooldown(guildId uint64) bool {
	key := fmt.Sprintf("synccooldown:%d", guildId)
	res, err := redis.Client.Exists(key).Result(); if err != nil {
		return true
	}

	return res == 1
}

func (s SyncCommand) addCooldown(guildId uint64) {
	key := fmt.Sprintf("synccooldown:%d", guildId)
	redis.Client.Set(key, "1", cooldown)
}

func processDeletedTickets(ctx command.CommandContext) (updated int) {
	tickets, err := dbclient.Client.Tickets.GetGuildOpenTickets(ctx.GuildId())
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	for _, ticket := range tickets {
		if ticket.ChannelId == nil {
			continue
		}

		_, err := ctx.Worker().GetChannel(*ticket.ChannelId)
		if err != nil { // An admin has deleted the channel manually
			updated++

			go func() {
				if err := dbclient.Client.Tickets.Close(ticket.Id, ticket.GuildId); err != nil {
					sentry.ErrorWithContext(err, ctx.ToErrorContext())
				}
			}()
		}
	}

	return
}

func processDeletedPanels(ctx command.CommandContext) (removed int) {
	panels, err := dbclient.Client.Panel.GetByGuild(ctx.GuildId())
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	for _, panel := range panels {
		// Pre-channel ID logging panel - we'll just leave it for now.
		if panel.ChannelId == 0 {
			continue
		}

		// Check cache first to prevent extra requests to discord
		if _, err := ctx.Worker().GetChannelMessage(panel.ChannelId, panel.MessageId); err != nil {
			removed++

			// Message no longer exists
			go func() {
				if err := dbclient.Client.Panel.Delete(panel.MessageId); err != nil {
					sentry.ErrorWithContext(err, ctx.ToErrorContext())
				}
			}()
		}
	}

	return
}

