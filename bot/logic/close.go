package logic

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/rest"
	"github.com/rxdn/gdl/rest/request"
	"strconv"
	"time"
)

func CloseTicket(ctx registry.CommandContext, reason *string, fromInteraction bool) {
	var success bool
	errorContext := ctx.ToErrorContext()

	// Get ticket struct
	ticket, err := dbclient.Client.Tickets.GetByChannel(ctx.ChannelId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	isTicket := ticket.GuildId != 0

	if !isTicket {
		if !fromInteraction {
			ctx.Reply(utils.Red, "Error", i18n.MessageNotATicketChannel)
			ctx.Reject()
		}

		return
	}

	defer func() {
		if !success {
			if err := dbclient.Client.AutoCloseExclude.Exclude(ticket.GuildId, ticket.Id); err != nil {
				sentry.ErrorWithContext(err, errorContext)
			}
		}
	}()

	// Check the user is permitted to close the ticket
	usersCanClose, err := dbclient.Client.UsersCanClose.Get(ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	permissionLevel, err := ctx.UserPermissionLevel()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	member, err := ctx.Member()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if permissionLevel == permission.Everyone && (ticket.UserId != member.User.Id || !usersCanClose) {
		if !fromInteraction {
			ctx.Reply(utils.Red, "Error", i18n.MessageCloseNoPermission)
			ctx.Reject()
		}
		return
	}

	// TODO: Re-add permission check
	/*if !permission.HasPermissions(s, guildId, s.SelfId(), permission.ManageChannels) {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "I do not have permission to delete this channel")
		return
	}*/

	if !fromInteraction {
		ctx.Accept()
	}

	// Archive
	msgs := make([]message.Message, 0)

	lastId := uint64(0)
	count := -1
	for count != 0 {
		array, err := ctx.Worker().GetChannelMessages(ctx.ChannelId(), rest.GetChannelMessagesData{
			Before: lastId,
			Limit:  100,
		})

		count = len(array)
		if err != nil {
			count = 0
			sentry.ErrorWithContext(err, errorContext)

			// First rest interaction, check for 403
			if err, ok := err.(request.RestError); ok && err.StatusCode == 403 {
				if err := dbclient.Client.AutoCloseExclude.ExcludeAll(ctx.GuildId()); err != nil {
					sentry.ErrorWithContext(err, errorContext)
				}
			}
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

	if err := utils.ArchiverClient.Store(msgs, ctx.GuildId(), ticket.Id, ctx.PremiumTier() > premium.None); err != nil {
		sentry.ErrorWithContext(err, errorContext)
	}

	// Set ticket state as closed and delete channel
	if err := dbclient.Client.Tickets.Close(ticket.Id, ctx.GuildId()); err != nil {
		ctx.HandleError(err)
		return
	}

	success = true

	// set close reason
	if reason != nil {
		if err := dbclient.Client.CloseReason.Set(ctx.GuildId(), ticket.Id, *reason); err != nil {
			ctx.HandleError(err)
			return
		}
	}

	if _, err := ctx.Worker().DeleteChannel(ctx.ChannelId()); err != nil {
		// Check if we should exclude this from autoclose
		if restError, ok := err.(request.RestError); ok && restError.StatusCode == 403 {
			if err := dbclient.Client.AutoCloseExclude.Exclude(ticket.GuildId, ticket.Id); err != nil {
				sentry.ErrorWithContext(err, errorContext)
			}
		}

		ctx.HandleError(err)
		return
	}

	// Save space - delete the webhook
	go dbclient.Client.Webhooks.Delete(ctx.GuildId(), ticket.Id)

	sendCloseEmbed(ctx, errorContext, member, ticket, reason)
}

func sendCloseEmbed(ctx registry.CommandContext, errorContext sentry.ErrorContext, member member.Member, ticket database.Ticket, reason *string) {
	// Send logs to archive channel
	archiveChannelId, err := dbclient.Client.ArchiveChannel.Get(ticket.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
	}

	var archiveChannelExists bool
	if archiveChannelId != 0 {
		if _, err := ctx.Worker().GetChannel(archiveChannelId); err == nil {
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
		SetTimestamp(time.Now()).
		AddField("Ticket ID", strconv.Itoa(ticket.Id), true).
		AddField("Opened By", fmt.Sprintf("<@%d>", ticket.UserId), true).
		AddField("Closed By", member.User.Mention(), true).
		AddField("Reason", formattedReason, false).
		AddField("Archive", fmt.Sprintf("[Click here](https://panel.ticketsbot.net/manage/%d/logs/view/%d)", ticket.GuildId, ticket.Id), true).
		AddField("Open Time", message.BuildTimestamp(ticket.OpenTime, message.TimestampStyleShortDateTime), true).
		AddField("Claimed By", claimedBy, true)

	if archiveChannelExists {
		if _, err := ctx.Worker().CreateMessageEmbed(archiveChannelId, embed); err != nil {
			sentry.ErrorWithContext(err, errorContext)
		}
	}

	// Notify user and send logs in DMs
	dmChannel, ok := getDmChannel(ctx, ticket.UserId)
	if ok {
		if _, err := ctx.Worker().CreateMessageEmbed(dmChannel, embed); err != nil {
			sentry.ErrorWithContext(err, errorContext)
		}
	}
}

func getDmChannel(ctx registry.CommandContext, userId uint64) (uint64, bool) {
	// Hack for autoclose
	if ctx.Worker().BotId == userId {
		return 0, false
	}

	cachedId, err := redis.GetDMChannel(userId)
	if err != nil { // We can continue
		if err != redis.ErrNotCached {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		}
	} else { // We have it cached
		if cachedId == nil {
			return 0, false
		} else {
			return *cachedId, true
		}
	}

	ch, err := ctx.Worker().CreateDM(userId)
	if err != nil {
		// check for 403
		if err, ok := err.(request.RestError); ok && err.StatusCode == 403 {
			if err := redis.StoreNullDMChannel(userId); err != nil {
				sentry.ErrorWithContext(err, ctx.ToErrorContext())
			}

			return 0, false
		}

		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return 0, false
	}

	if err := redis.StoreDMChannel(userId, ch.Id); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	return ch.Id, true
}
