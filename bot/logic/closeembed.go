package logic

import (
	"fmt"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild/emoji"
	"github.com/rxdn/gdl/objects/interaction/component"
	"github.com/rxdn/gdl/rest"
	"strconv"
	"time"
)

type CloseEmbedElement func(ctx registry.CommandContext, ticket database.Ticket) []component.Component

func NoopElement() CloseEmbedElement {
	return func(ctx registry.CommandContext, ticket database.Ticket) []component.Component {
		return nil
	}
}

func TranscriptLinkElement(condition bool) CloseEmbedElement {
	if !condition {
		return NoopElement()
	}

	return func(ctx registry.CommandContext, ticket database.Ticket) []component.Component {
		var transcriptEmoji *emoji.Emoji
		if !ctx.Worker().IsWhitelabel {
			transcriptEmoji = customisation.EmojiTranscript.BuildEmoji()
		}

		transcriptLink := fmt.Sprintf("https://panel.ticketsbot.net/manage/%d/transcripts/view/%d", ticket.GuildId, ticket.Id)

		return utils.Slice(component.BuildButton(component.Button{
			Label: "View Online Transcript",
			Style: component.ButtonStyleLink,
			Emoji: transcriptEmoji,
			Url:   utils.Ptr(transcriptLink),
		}))
	}
}

func ThreadLinkElement(condition bool) CloseEmbedElement {
	if !condition {
		return NoopElement()
	}

	return func(ctx registry.CommandContext, ticket database.Ticket) []component.Component {
		var threadEmoji *emoji.Emoji
		if !ctx.Worker().IsWhitelabel {
			threadEmoji = customisation.EmojiThread.BuildEmoji()
		}

		return utils.Slice(
			component.BuildButton(component.Button{
				Label: "View Thread",
				Style: component.ButtonStyleLink,
				Emoji: threadEmoji,
				Url:   utils.Ptr(fmt.Sprintf("https://discord.com/channels/%d/%d", ticket.GuildId, *ticket.ChannelId)),
			}),
		)
	}
}

func ViewFeedbackElement(condition bool) CloseEmbedElement {
	if !condition {
		return NoopElement()
	}

	return func(ctx registry.CommandContext, ticket database.Ticket) []component.Component {
		return utils.Slice(
			component.BuildButton(component.Button{
				Label:    "View Exit Survey",
				CustomId: fmt.Sprintf("view-survey-%d-%d", ticket.GuildId, ticket.Id),
				Style:    component.ButtonStylePrimary,
				Emoji:    utils.BuildEmoji("📰"),
			}),
		)
	}
}

func FeedbackRowElement(condition bool) CloseEmbedElement {
	if !condition {
		return NoopElement()
	}

	return func(ctx registry.CommandContext, ticket database.Ticket) []component.Component {
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

		return buttons
	}
}

func BuildCloseEmbed(
	ctx registry.CommandContext,
	ticket database.Ticket,
	closedBy uint64,
	reason *string,
	rating *uint8,
	components [][]CloseEmbedElement,
) (*embed.Embed, []component.Component) {
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
		AddField(formatTitle("Ticket ID", customisation.EmojiId, ctx.Worker().IsWhitelabel), strconv.Itoa(ticket.Id), true).
		AddField(formatTitle("Opened By", customisation.EmojiOpen, ctx.Worker().IsWhitelabel), fmt.Sprintf("<@%d>", ticket.UserId), true).
		AddField(formatTitle("Closed By", customisation.EmojiClose, ctx.Worker().IsWhitelabel), fmt.Sprintf("<@%d>", closedBy), true).
		AddField(formatTitle("Open Time", customisation.EmojiOpenTime, ctx.Worker().IsWhitelabel), message.BuildTimestamp(ticket.OpenTime, message.TimestampStyleShortDateTime), true).
		AddField(formatTitle("Claimed By", customisation.EmojiClaim, ctx.Worker().IsWhitelabel), claimedBy, true)

	if rating == nil {
		closeEmbed = closeEmbed.AddBlankField(true)
	} else {
		closeEmbed = closeEmbed.AddField(formatTitle("Rating", customisation.EmojiRating, ctx.Worker().IsWhitelabel), fmt.Sprintf("%d ⭐", *rating), true)
	}

	closeEmbed = closeEmbed.AddField(formatTitle("Reason", customisation.EmojiReason, ctx.Worker().IsWhitelabel), formattedReason, false)

	var rows []component.Component
	for _, row := range components {
		var rowElements []component.Component
		for _, element := range row {
			rowElements = append(rowElements, element(ctx, ticket)...)
		}

		if len(rowElements) > 0 {
			rows = append(rows, component.BuildActionRow(rowElements...))
		}
	}

	return closeEmbed, rows
}

func formatTitle(s string, emoji customisation.CustomEmoji, isWhitelabel bool) string {
	if !isWhitelabel {
		return fmt.Sprintf("%s %s", emoji, s)
	} else {
		return s
	}
}

func EditGuildArchiveMessageIfExists(
	ctx registry.CommandContext,
	ticket database.Ticket,
	settings database.Settings,
	viewFeedbackButton bool,
	closedBy uint64,
	reason *string,
	rating *uint8,
) error {
	archiveMessage, ok, err := dbclient.Client.ArchiveMessages.Get(ticket.GuildId, ticket.Id)
	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	componentBuilders := [][]CloseEmbedElement{
		{
			TranscriptLinkElement(settings.StoreTranscripts),
			ThreadLinkElement(ticket.IsThread && ticket.ChannelId != nil),
			ViewFeedbackElement(viewFeedbackButton),
		},
	}

	embed, components := BuildCloseEmbed(ctx, ticket, closedBy, reason, rating, componentBuilders)
	_, err = ctx.Worker().EditMessage(archiveMessage.ChannelId, archiveMessage.MessageId, rest.EditMessageData{
		Embeds:     utils.Slice(embed),
		Components: components,
	})

	return err
}
