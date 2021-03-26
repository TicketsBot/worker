package logic

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/rest"
	"github.com/rxdn/gdl/rest/request"
	"strconv"
	"time"
)

func CloseTicket(worker *worker.Context, guildId, channelId, messageId uint64, member member.Member, reason *string, fromReaction, isPremium bool) {
	errorContext := errorcontext.WorkerErrorContext{
		Guild:   guildId,
		User:    member.User.Id,
		Channel: channelId,
		Shard:   worker.ShardId,
	}

	replyTo := utils.CreateReference(messageId, channelId, guildId)

	// Get ticket struct
	ticket, err := dbclient.Client.Tickets.GetByChannel(channelId)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	isTicket := ticket.GuildId != 0

	// Cannot happen if fromReaction
	if !isTicket {
		if !fromReaction {
			utils.ReactWithCross(worker, channelId, messageId)
			utils.SendEmbed(worker, channelId, guildId, replyTo, utils.Red, "Error", translations.MessageNotATicketChannel, nil, 30, isPremium)
		}

		return
	}

	// Check the user is permitted to close the ticket
	permissionLevel, err := permission.GetPermissionLevel(utils.ToRetriever(worker), member, guildId)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	usersCanClose, err := dbclient.Client.UsersCanClose.Get(guildId)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	if (permissionLevel == permission.Everyone && ticket.UserId != member.User.Id) || (permissionLevel == permission.Everyone && !usersCanClose) {
		if !fromReaction {
			utils.ReactWithCross(worker, channelId, messageId)
			utils.SendEmbed(worker, channelId, guildId, replyTo, utils.Red, "Error", translations.MessageCloseNoPermission, nil, 30, isPremium)
		}
		return
	}

	// TODO: Re-add permission check
	/*if !permission.HasPermissions(s, guildId, s.SelfId(), permission.ManageChannels) {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "I do not have permission to delete this channel")
		return
	}*/

	if !fromReaction {
		utils.ReactWithCheck(worker, channelId, messageId)
	}

	// Archive
	msgs := make([]message.Message, 0)

	lastId := uint64(0)
	count := -1
	for count != 0 {
		array, err := worker.GetChannelMessages(channelId, rest.GetChannelMessagesData{
			Before: lastId,
			Limit:  100,
		})

		count = len(array)
		if err != nil {
			count = 0
			sentry.Error(err)
		}

		if count > 0 {
			lastId = array[len(array)-1].Id

			msgs = append(msgs, array...)
		}
	}

	// Reverse messages
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	if err := utils.ArchiverClient.Store(msgs, guildId, ticket.Id, isPremium); err != nil {
		sentry.ErrorWithContext(err, errorContext)
	}

	// Set ticket state as closed and delete channel
	if err := dbclient.Client.Tickets.Close(ticket.Id, guildId); err != nil {
		sentry.ErrorWithContext(err, errorContext)
	}

	// set close reason
	if reason != nil {
		if err := dbclient.Client.CloseReason.Set(guildId, ticket.Id, *reason); err != nil {
			sentry.ErrorWithContext(err, errorContext)
		}
	}

	if _, err := worker.DeleteChannel(channelId); err != nil {
		// Check if we should exclude this from autoclose
		if restError, ok := err.(request.RestError); ok && restError.StatusCode == 403 {
			if err := dbclient.Client.AutoCloseExclude.Exclude(ticket.GuildId, ticket.Id); err != nil {
				sentry.ErrorWithContext(err, errorContext)
			}
		}

		sentry.ErrorWithContext(err, errorContext)
	}

	// Save space - delete the webhook
	go dbclient.Client.Webhooks.Delete(guildId, ticket.Id)

	sendCloseEmbed(worker, errorContext, member, ticket, reason)
}

func sendCloseEmbed(worker *worker.Context, errorContext sentry.ErrorContext, member member.Member, ticket database.Ticket, reason *string) {
	// Send logs to archive channel
	archiveChannelId, err := dbclient.Client.ArchiveChannel.Get(ticket.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
	}

	var archiveChannelExists bool
	if archiveChannelId != 0 {
		if _, err := worker.GetChannel(archiveChannelId); err == nil {
			archiveChannelExists = true
		}
	}

	var formattedReason string
	if reason == nil {
		formattedReason = "No reason specified"
	} else {
		formattedReason = *reason
		if len(formattedReason) > 1024 {
			formattedReason = formattedReason[:1024]
		}
	}

	var claimedBy string
	{
		claimUserId, err := dbclient.Client.TicketClaims.Get(ticket.GuildId, ticket.Id)
		if err != nil {
			sentry.Error(err)
		}

		if claimUserId == 0 {
			claimedBy = "Not claimed"
		} else {
			claimedBy = fmt.Sprintf("<@%d>", claimUserId)
		}
	}

	embed := embed.NewEmbed().
		SetTitle("Ticket Closed").
		SetColor(int(utils.Green)).
		AddField("Ticket ID", strconv.Itoa(ticket.Id), true).
		AddField("Opened By", fmt.Sprintf("<@%d>", ticket.UserId), true).
		AddField("Closed By", member.User.Mention(), true).
		AddField("Reason", formattedReason, false).
		AddField("Archive", fmt.Sprintf("[Click here](https://panel.ticketsbot.net/manage/%d/logs/view/%d)", ticket.GuildId, ticket.Id), true).
		AddField("Open Time", utils.FormatDateTime(ticket.OpenTime), true).
		AddField("Claimed By", claimedBy, true).
		SetFooter(fmt.Sprintf("Close Time: %s", utils.FormatDateTime(time.Now())), "")

	if archiveChannelExists {
		if _, err := worker.CreateMessageEmbed(archiveChannelId, embed); err != nil {
			sentry.ErrorWithContext(err, errorContext)
		}
	}

	// Notify user and send logs in DMs
	if dmChannel, err := worker.CreateDM(ticket.UserId); err == nil {
		// Only send the msg if we could create the channel
		if _, err := worker.CreateMessageEmbed(dmChannel.Id, embed); err != nil {
			sentry.ErrorWithContext(err, errorContext)
		}
	}
}
