package listeners

import (
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/rxdn/gdl/objects/interaction"
	"regexp"
	"strconv"
)

var pattern = regexp.MustCompile(`rate_(\d+)_(\d+)_([1-5])`)

// TODO: Proper context
func OnRate(worker *worker.Context, data interaction.ButtonInteraction) {
	groups := pattern.FindStringSubmatch(data.Data.CustomId)
	if len(groups) < 4 {
		return
	}

	// Errors are impossible
	guildId, _ := strconv.ParseUint(groups[1], 10, 64)
	ticketId, _  := strconv.Atoi(groups[2])
	ratingRaw, _ := strconv.Atoi(groups[3])
	rating := uint8(ratingRaw)

	// DMs only, so data.User is not null
	if data.User == nil {
		return
	}

	errorCtx := errorcontext.WorkerErrorContext{
		Guild:   guildId,
		User:    data.User.Id,
		Channel: data.ChannelId,
	}

	// Get ticket
	ticket, err := dbclient.Client.Tickets.Get(ticketId, guildId)
	if err != nil {
		sentry.ErrorWithContext(err, errorCtx)
		return
	}

	if ticket.UserId != data.User.Id || ticket.GuildId != guildId || ticket.Id != ticketId {
		return
	}

	feedbackEnabled, err := dbclient.Client.FeedbackEnabled.Get(guildId)
	if err != nil {
		sentry.ErrorWithContext(err, errorCtx)
		return
	}

	if !feedbackEnabled {
		if _, err := worker.CreateMessage(data.ChannelId, "This server has feedback disabled"); err != nil {
			sentry.ErrorWithContext(err, errorCtx)
		}

		return
	}

	if err := dbclient.Client.ServiceRatings.Set(guildId, ticketId, rating); err != nil {
		sentry.ErrorWithContext(err, errorCtx)
		return
	}

	if _, err := worker.CreateMessage(data.ChannelId, "Your feedback has been recorded"); err != nil {
		sentry.ErrorWithContext(err, errorCtx)
	}
}
