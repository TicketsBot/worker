package logic

import (
	"fmt"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/metrics/statsd"
	"github.com/TicketsBot/worker/bot/redis"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/rest"
	"github.com/rxdn/gdl/rest/request"
	"time"
)

func CloseTicket(ctx registry.CommandContext, reason *string, bypassPermissionCheck bool) {
	var success bool
	errorContext := ctx.ToErrorContext()

	// Get ticket struct
	ticket, err := dbclient.Client.Tickets.GetByChannelAndGuild(ctx.ChannelId(), ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if ticket.Id == 0 || ticket.GuildId != ctx.GuildId() {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageNotATicketChannel)
		return
	}

	defer func() {
		if !success {
			if err := dbclient.Client.AutoCloseExclude.Exclude(ticket.GuildId, ticket.Id); err != nil {
				sentry.ErrorWithContext(err, errorContext)
			}
		}
	}()

	if !bypassPermissionCheck && !utils.CanClose(ctx, ticket) {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageCloseNoPermission)
		return
	}

	member, err := ctx.Member()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	settings, err := ctx.Settings()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Archive
	if settings.StoreTranscripts {
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

				break
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

		err := utils.ArchiverClient.Store(msgs, ctx.GuildId(), ticket.Id, ctx.PremiumTier() > premium.None)
		if err == nil {
			if err := dbclient.Client.Tickets.SetHasTranscript(ctx.GuildId(), ticket.Id, true); err != nil {
				sentry.ErrorWithContext(err, errorContext)
			}
		} else {
			sentry.ErrorWithContext(err, errorContext)
		}
	}

	// Set ticket state as closed and delete channel
	if err := dbclient.Client.Tickets.Close(ticket.Id, ctx.GuildId()); err != nil {
		ctx.HandleError(err)
		return
	}

	success = true

	// set close reason + user
	closeMetadata := database.CloseMetadata{
		Reason: reason,
	}

	if member.User.Id != ctx.UserId() {
		closeMetadata.ClosedBy = &member.User.Id
	}

	if err := dbclient.Client.CloseReason.Set(ctx.GuildId(), ticket.Id, closeMetadata); err != nil {
		ctx.HandleError(err)
		return
	}

	if ticket.IsThread {
		// If it is a thread, we need to send a message
		if reason == nil {
			ctx.ReplyPermanent(customisation.Green, i18n.TitleTicketClosed, i18n.MessageCloseSuccess, ctx.UserId())
		} else {
			fields := []embed.EmbedField{
				{
					Name:   ctx.GetMessage(i18n.Reason),
					Value:  fmt.Sprintf("```%s```", *reason),
					Inline: false,
				},
			}

			ctx.ReplyWithFieldsPermanent(customisation.Green, i18n.TitleTicketClosed, i18n.MessageCloseSuccess, fields, ctx.UserId())
		}

		// Discord has a race condition
		time.Sleep(time.Millisecond * 500)

		data := rest.ModifyChannelData{
			ThreadMetadataModifyData: &rest.ThreadMetadataModifyData{
				Archived: utils.Ptr(true),
				Locked:   utils.Ptr(true),
			},
		}

		if _, err := ctx.Worker().ModifyChannel(ctx.ChannelId(), data); err != nil {
			ctx.HandleError(err)
			return
		}
	} else {
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
	}

	// Save space - delete the webhook
	if !ticket.IsThread {
		go dbclient.Client.Webhooks.Delete(ctx.GuildId(), ticket.Id)
	}

	if err := dbclient.Client.CloseRequest.Delete(ticket.GuildId, ticket.Id); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	// Delete join thread button
	if ticket.IsThread && ticket.JoinMessageId != nil && settings.TicketNotificationChannel != nil {
		_ = ctx.Worker().DeleteMessage(*settings.TicketNotificationChannel, *ticket.JoinMessageId)
		if err := dbclient.Client.Tickets.SetJoinMessageId(ticket.GuildId, ticket.Id, nil); err != nil {
			sentry.ErrorWithContext(err, errorContext)
		}
	}

	sendCloseEmbed(ctx, errorContext, member, settings, ticket, reason)
}

func sendCloseEmbed(ctx registry.CommandContext, errorContext sentry.ErrorContext, member member.Member, settings database.Settings, ticket database.Ticket, reason *string) {
	// Send logs to archive channel
	archiveChannelId, err := dbclient.Client.ArchiveChannel.Get(ticket.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
	}

	var archiveChannelExists bool
	if archiveChannelId != nil {
		if _, err := ctx.Worker().GetChannel(*archiveChannelId); err == nil {
			archiveChannelExists = true
		}
	}

	if archiveChannelExists && archiveChannelId != nil {
		componentBuilders := [][]CloseEmbedElement{
			{
				TranscriptLinkElement(settings.StoreTranscripts),
				ThreadLinkElement(ticket.IsThread && ticket.ChannelId != nil),
			},
		}

		closeEmbed, closeComponents := BuildCloseEmbed(ctx, ticket, member.User.Id, reason, nil, componentBuilders)

		data := rest.CreateMessageData{
			Embeds:     utils.Slice(closeEmbed),
			Components: closeComponents,
		}

		msg, err := ctx.Worker().CreateMessageComplex(*archiveChannelId, data)
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
		} else {
			// Add message to archive
			if err := dbclient.Client.ArchiveMessages.Set(ticket.GuildId, ticket.Id, *archiveChannelId, msg.Id); err != nil {
				ctx.HandleError(err)
				return
			}
		}
	}

	// Notify user and send logs in DMs
	// This mutates state!
	dmChannel, ok := getDmChannel(ctx, ticket.UserId)
	if ok {
		guild, err := ctx.Guild()
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
			return
		}

		feedbackEnabled, err := dbclient.Client.FeedbackEnabled.Get(ctx.GuildId())
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
			return
		}

		// Only offer to take feedback if the user has sent a message
		hasSentMessage, err := dbclient.Client.Participants.HasParticipated(ctx.GuildId(), ticket.Id, ticket.UserId)
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
			return
		}

		statsd.Client.IncrementKey(statsd.KeyDirectMessage)

		componentBuilders := [][]CloseEmbedElement{
			{
				TranscriptLinkElement(settings.StoreTranscripts),
				ThreadLinkElement(ticket.IsThread && ticket.ChannelId != nil),
			},
			{
				FeedbackRowElement(feedbackEnabled && hasSentMessage),
			},
		}

		closeEmbed, closeComponents := BuildCloseEmbed(ctx, ticket, member.User.Id, reason, nil, componentBuilders)
		closeEmbed.SetAuthor(guild.Name, "", fmt.Sprintf("https://cdn.discordapp.com/icons/%d/%s.png", guild.Id, guild.Icon))

		data := rest.CreateMessageData{
			Embeds:     utils.Slice(closeEmbed),
			Components: closeComponents,
		}

		if _, err := ctx.Worker().CreateMessageComplex(dmChannel, data); err != nil {
			sentry.ErrorWithContext(err, errorContext)
		}
	}
}

func getDmChannel(ctx registry.CommandContext, userId uint64) (uint64, bool) {
	// Hack for autoclose
	if ctx.Worker().BotId == userId {
		return 0, false
	}

	cachedId, err := redis.GetDMChannel(userId, ctx.Worker().BotId)
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
			if err := redis.StoreNullDMChannel(userId, ctx.Worker().BotId); err != nil {
				sentry.ErrorWithContext(err, ctx.ToErrorContext())
			}

			return 0, false
		}

		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return 0, false
	}

	if err := redis.StoreDMChannel(userId, ch.Id, ctx.Worker().BotId); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	return ch.Id, true
}
