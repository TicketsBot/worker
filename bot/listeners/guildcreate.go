package listeners

import (
	"fmt"
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"github.com/rxdn/gdl/objects/auditlog"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/rest"
)

// Fires when we receive a guild
func OnGuildCreate(worker *worker.Context, e *events.GuildCreate, extra eventforwarding.Extra) {
	if worker.IsWhitelabel {
		if err := dbclient.Client.WhitelabelGuilds.Add(worker.BotId, e.Guild.Id); err != nil {
			sentry.Error(err)
		}
	}

	if extra.IsJoin {
		go statsd.IncrementKey(statsd.JOINS)

		//sendIntroMessage(worker, e.Guild, e.Guild.OwnerId)

		// find who invited the bot
		/*if inviter := getInviter(worker, e.Guild.Id); inviter != 0 && inviter != e.Guild.OwnerId {
			sendIntroMessage(worker, e.Guild, inviter)
		}*/
	}
}

func sendIntroMessage(worker *worker.Context, guild guild.Guild, userId uint64) {
	// Create DM channel
	channel, err := worker.CreateDM(userId)
	if err != nil { // User probably has DMs disabled
		return
	}

	msg := embed.NewEmbed().
		SetTitle("Tickets").
		SetDescription("Thank you for inviting Tickets to your server! Below is a quick guide on setting the bot up, please don't hesitate to contact us in our [support server](https://discord.gg/VtV3rSk) if you need any assistance!").
		SetColor(int(utils.Green)).
		AddField("Setup", "You can setup the bot using `t!setup`, or you can use the [dashboard](https://panel.ticketsbot.net) which has additional options", false).
		AddField("Reaction Panels", fmt.Sprintf("Reaction panels are a commonly used feature of the bot. You can read about them [here](https://ticketsbot.net/panels), or create one on [the dashboard](https://panel.ticketsbot.net/manage/%d/panels)", guild.Id), false).
		AddField("Adding Staff", "To make staff able to answer tickets, you must let the bot know about them first. You can do this through\n`t!addsupport [@User / @Role]` and `t!addadmin [@User / @Role]`. Administrators can change the settings of the bot and access the dashboard.", false).
		AddField("Tags", "Tags are predefined tickets of text which you can access through a simple command. You can learn more about them [here](https://ticketsbot.net/tags).", false).
		AddField("Claiming", "Tickets can be claimed by your staff such that other staff members cannot also reply to the ticket. You can learn more about claiming [here](https://ticketsbot.net/claiming).", false).
		AddField("Additional Support", "If you are still confused, we welcome you to our [support server](https://discord.gg/VtV3rSk). Cheers.", false)

	_, _ = worker.CreateMessageEmbed(channel.Id, msg)
}

func getInviter(worker *worker.Context, guildId uint64) (userId uint64) {
	data := rest.GetGuildAuditLogData{
		ActionType: auditlog.EventBotAdd,
		Limit:      50,
	}

	auditLog, err :=worker.GetGuildAuditLog(guildId, data)
	if err != nil {
		sentry.Error(err) // prob perms
		return
	}

	for _, entry := range auditLog.Entries {
		if entry.ActionType != auditlog.EventBotAdd || entry.TargetId != worker.BotId {
			continue
		}

		userId = entry.UserId
		break
	}

	return
}
