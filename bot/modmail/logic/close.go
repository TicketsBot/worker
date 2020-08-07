package logic

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/rest"
	"strings"
	"time"
)

func HandleClose(session database.ModmailSession, ctx command.CommandContext) {
	reason := strings.Join(ctx.Args, " ")

	// Check the user is permitted to close the ticket
	if ctx.UserPermissionLevel == permission.Everyone && session.UserId != ctx.Author.Id {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", translations.MessageCloseNoPermission)
		return
	}

	// TODO: Re-add permission check
	/*if !permission.HasPermissions(ctx.Shard, ctx.GuildId, ctx.Shard.SelfId(), permission.ManageChannels) {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "I do not have permission to delete this channel")
		return
	}*/

	if ctx.ShouldReact {
		ctx.ReactWithCheck()
	}

	// Archive
	msgs := make([]message.Message, 0)

	lastId := uint64(0)
	count := -1
	for count != 0 {
		array, err := ctx.Worker.GetChannelMessages(ctx.ChannelId, rest.GetChannelMessagesData{
			Before: lastId,
			Limit:  100,
		})

		count = len(array)
		if err != nil {
			count = 0
			sentry.LogWithContext(err, ctx.ToErrorContext())
		}

		if count > 0 {
			lastId = array[len(array)-1].Id

			for _, msg := range array {
				msgs = append(msgs, msg)
				if msg.Id == session.WelcomeMessageId {
					count = 0
					break
				}
			}
		}
	}

	// Reverse messages
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	// we don't use this yet so chuck it in a goroutine
	go func() {
		isPremium := utils.PremiumClient.GetTierByGuildId(ctx.GuildId, true, ctx.Worker.Token, ctx.Worker.RateLimiter) > premium.None
		if err := utils.ArchiverClient.StoreModmail(msgs, session.GuildId, session.Uuid.String(), isPremium); err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		}
	}()

	// Delete the webhook
	// We need to block for this
	if err := dbclient.Client.ModmailWebhook.Delete(session.Uuid); err != nil {
		ctx.HandleError(err)
		return
	}

	// Set ticket state as closed and delete channel
	go dbclient.Client.ModmailSession.DeleteByUser(ctx.Worker.BotId, session.UserId)
	if _, err := ctx.Worker.DeleteChannel(session.StaffChannelId); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	// Send logs to archive channel
	archiveChannelId, err := dbclient.Client.ArchiveChannel.Get(session.GuildId); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	var channelExists bool
	if archiveChannelId != 0 {
		if _, err := ctx.Worker.GetChannel(archiveChannelId); err == nil {
			channelExists = true
		}
	}

	if channelExists {
		embed := embed.NewEmbed().
			SetTitle("Ticket Closed").
			SetColor(int(utils.Green)).
			AddField("Closed By", ctx.Author.Mention(), true).
			AddField("Archive", fmt.Sprintf("[Click here](https://panel.ticketsbot.net/manage/%d/logs/modmail/view/%s)", session.GuildId, session.Uuid), true)

		if reason == "" {
			embed.AddField("Reason", "No reason specified", false)
		} else {
			embed.AddField("Reason", reason, false)
		}

		if _, err := ctx.Worker.CreateMessageEmbed(archiveChannelId, embed); err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		}
	}

	go dbclient.Client.ModmailArchive.Set(database.ModmailArchive{
		Uuid:      session.Uuid,
		GuildId:   session.GuildId,
		UserId:    session.UserId,
		CloseTime: time.Now(),
	})

	guild, err := ctx.Worker.GetGuild(session.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	// Notify user and send logs in DMs
	privateMessage, err := ctx.Worker.CreateDM(session.UserId)
	if err == nil {
		var content string
		// Create message content
		if ctx.Author.Id == session.UserId {
			content = fmt.Sprintf("You closed your modmail ticket in `%s`", guild.Name)
		} else if len(ctx.Args) == 0 {
			content = fmt.Sprintf("Your modmail ticket in `%s` was closed by %s", guild.Name, ctx.Author.Mention())
		} else {
			content = fmt.Sprintf("Your modmail ticket in `%s` was closed by %s with reason `%s`", guild.Name, ctx.Author.Mention(), reason)
		}

		// Errors occur when users have privacy settings high
		_, _ = ctx.Worker.CreateMessage(privateMessage.Id, content)
	}
}

