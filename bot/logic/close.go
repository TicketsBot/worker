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
	"github.com/rxdn/gdl/objects"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild/emoji"
	"github.com/rxdn/gdl/objects/interaction/component"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/rest"
	"github.com/rxdn/gdl/rest/request"
	"strconv"
	"time"
)

func CloseTicket(ctx registry.CommandContext, reason *string) {
	var success bool
	errorContext := ctx.ToErrorContext()

	// Get ticket struct
	ticket, err := dbclient.Client.Tickets.GetByChannelAndGuild(ctx.ChannelId(), ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	isTicket := ticket.GuildId != 0

	if !isTicket {
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

	if !utils.CanClose(ctx, ticket) {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageCloseNoPermission)
		return
	}

	member, err := ctx.Member()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	settings, err := dbclient.Client.Settings.Get(ctx.GuildId())
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

	if err := dbclient.Client.CloseRequest.Delete(ticket.GuildId, ticket.Id); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
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

	closeEmbed, closeComponents := buildCloseEmbed(ctx, ticket, settings, member, reason)

	if archiveChannelExists && archiveChannelId != nil {
		data := rest.CreateMessageData{
			Embeds:     utils.Slice(closeEmbed),
			Components: closeComponents,
		}

		if _, err := ctx.Worker().CreateMessageComplex(*archiveChannelId, data); err != nil {
			sentry.ErrorWithContext(err, errorContext)
		}
	}

	// Notify user and send logs in DMs
	dmChannel, ok := getDmChannel(ctx, ticket.UserId)
	if ok {
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

		go statsd.Client.IncrementKey(statsd.KeyDirectMessage)

		if !feedbackEnabled || !hasSentMessage {
			data := rest.CreateMessageData{
				Embeds:     []*embed.Embed{closeEmbed},
				Components: append(closeComponents),
			}

			if _, err := ctx.Worker().CreateMessageComplex(dmChannel, data); err != nil {
				sentry.ErrorWithContext(err, errorContext)
			}
		} else {
			closeEmbed.SetDescription("Please rate the quality of service received with the buttons below")

			data := rest.CreateMessageData{
				Embeds:     []*embed.Embed{closeEmbed},
				Components: append(closeComponents, buildRatingActionRow(ticket)),
			}

			if _, err := ctx.Worker().CreateMessageComplex(dmChannel, data); err != nil {
				sentry.ErrorWithContext(err, errorContext)
			}
		}
	}
}

func buildCloseEmbed(ctx registry.CommandContext, ticket database.Ticket, settings database.Settings, member member.Member, reason *string) (*embed.Embed, []component.Component) {
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

	// TODO: Translate titles
	closeEmbed := embed.NewEmbed().
		SetTitle("Ticket Closed").
		SetColor(ctx.GetColour(customisation.Green)).
		SetTimestamp(time.Now()).
		AddField(formatTitle("Ticket ID", utils.EmojiId, ctx.Worker().IsWhitelabel), strconv.Itoa(ticket.Id), true).
		AddField(formatTitle("Opened By", utils.EmojiOpen, ctx.Worker().IsWhitelabel), fmt.Sprintf("<@%d>", ticket.UserId), true).
		AddField(formatTitle("Closed By", utils.EmojiClose, ctx.Worker().IsWhitelabel), member.User.Mention(), true).
		AddField(formatTitle("Reason", utils.EmojiReason, ctx.Worker().IsWhitelabel), formattedReason, false)

	var components []component.Component
	if settings.StoreTranscripts {
		title := formatTitle("Transcript", utils.EmojiTranscript, ctx.Worker().IsWhitelabel)
		transcriptLink := fmt.Sprintf("https://panel.ticketsbot.net/manage/%d/transcripts/view/%d", ticket.GuildId, ticket.Id)
		closeEmbed.AddField(title, fmt.Sprintf("[Click here](%s)", transcriptLink), true)

		var transcriptEmoji *emoji.Emoji
		if !ctx.Worker().IsWhitelabel {
			transcriptEmoji = &emoji.Emoji{
				Id:   objects.NewNullableSnowflake(utils.EmojiTranscript.Id),
				Name: utils.EmojiTranscript.Name,
			}
		}

		components = []component.Component{
			component.BuildActionRow(
				component.BuildButton(component.Button{
					Label: "View Transcript",
					Style: component.ButtonStyleLink,
					Emoji: transcriptEmoji,
					Url:   utils.Ptr(transcriptLink),
				}),
			),
		}
	}

	closeEmbed.
		AddField(formatTitle("Open Time", utils.EmojiTime, ctx.Worker().IsWhitelabel), message.BuildTimestamp(ticket.OpenTime, message.TimestampStyleShortDateTime), true).
		AddField(formatTitle("Claimed By", utils.EmojiClaim, ctx.Worker().IsWhitelabel), claimedBy, true)

	return closeEmbed, components
}

func formatTitle(s string, emoji utils.CustomEmoji, isWhitelabel bool) string {
	if !isWhitelabel {
		return fmt.Sprintf("%s %s", emoji, s)
	} else {
		return s
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

func buildRatingActionRow(ticket database.Ticket) component.Component {
	buttons := make([]component.Component, 5)

	for i := 1; i <= 5; i++ {
		var style component.ButtonStyle
		if i <= 2 {
			style = component.ButtonStyleDanger
		} else if i == 3 {
			style = component.ButtonStylePrimary
		} else {
			style = component.ButtonStyleSuccess
		}

		buttons[i-1] = component.BuildButton(component.Button{
			Label:    strconv.Itoa(i),
			CustomId: fmt.Sprintf("rate_%d_%d_%d", ticket.GuildId, ticket.Id, i),
			Style:    style,
			Emoji: &emoji.Emoji{
				Name: "⭐",
			},
		})
	}

	return component.BuildActionRow(buttons...)
}
