package statistics

import (
	"context"
	"fmt"
	"github.com/TicketsBot/analytics-client"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"golang.org/x/sync/errgroup"
	"strconv"
	"time"
)

type StatsServerCommand struct {
}

func (StatsServerCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:             "server",
		Description:      i18n.HelpStatsServer,
		Type:             interaction.ApplicationCommandTypeChatInput,
		PermissionLevel:  permission.Support,
		Category:         command.Statistics,
		PremiumOnly:      true,
		DefaultEphemeral: true,
	}
}

func (c StatsServerCommand) GetExecutor() interface{} {
	return c.Execute
}

func (StatsServerCommand) Execute(ctx registry.CommandContext) {
	span := sentry.StartSpan(context.Background(), "/stats server")
	defer span.Finish()

	group, _ := errgroup.WithContext(context.Background())

	var totalTickets, openTickets int

	// totalTickets
	group.Go(func() (err error) {
		span := sentry.StartSpan(span.Context(), "GetTotalTicketCount")
		defer span.Finish()

		totalTickets, err = dbclient.Client.Tickets.GetTotalTicketCount(ctx.GuildId())
		return
	})

	// openTickets
	group.Go(func() error {
		span := sentry.StartSpan(span.Context(), "GetGuildOpenTickets")
		defer span.Finish()

		tickets, err := dbclient.Client.Tickets.GetGuildOpenTickets(ctx.GuildId())
		openTickets = len(tickets)
		return err
	})

	var feedbackRating float32
	var feedbackCount int

	group.Go(func() (err error) {
		span := sentry.StartSpan(span.Context(), "GetAverageFeedbackRating")
		defer span.Finish()

		feedbackRating, err = dbclient.Client.ServiceRatings.GetAverage(ctx.GuildId())
		return
	})

	group.Go(func() (err error) {
		span := sentry.StartSpan(span.Context(), "GetFeedbackCount")
		defer span.Finish()

		feedbackCount, err = dbclient.Client.ServiceRatings.GetCount(ctx.GuildId())
		return
	})

	// first response times
	var firstResponseTime analytics.TripleWindow
	group.Go(func() (err error) {
		span := sentry.StartSpan(span.Context(), "GetFirstResponseTimeStats")
		defer span.Finish()

		context, cancel := utils.ContextTimeout(time.Minute)
		defer cancel()

		firstResponseTime, err = dbclient.Analytics.GetFirstResponseTimeStats(context, ctx.GuildId())
		return
	})

	// ticket duration
	var ticketDuration analytics.TripleWindow
	group.Go(func() (err error) {
		span := sentry.StartSpan(span.Context(), "GetTicketDurationStats")
		defer span.Finish()

		context, cancel := utils.ContextTimeout(time.Minute)
		defer cancel()

		ticketDuration, err = dbclient.Analytics.GetTicketDurationStats(context, ctx.GuildId())
		return
	})

	if err := group.Wait(); err != nil {
		ctx.HandleError(err)
		return
	}

	span = sentry.StartSpan(span.Context(), "Send Message")

	msgEmbed := embed.NewEmbed().
		SetTitle("Statistics").
		SetColor(ctx.GetColour(customisation.Green)).
		AddField("Total Tickets", strconv.Itoa(totalTickets), true).
		AddField("Open Tickets", strconv.Itoa(openTickets), true).
		AddBlankField(true).
		AddField("Feedback Rating", fmt.Sprintf("%.1f / 5 ⭐", feedbackRating), true).
		AddField("Feedback Count", fmt.Sprintf("%d", feedbackCount), true).
		AddBlankField(true).
		AddField("Average First Response Time (Total)", formatNullableTime(firstResponseTime.AllTime), true).
		AddField("Average First Response Time (Monthly)", formatNullableTime(firstResponseTime.Monthly), true).
		AddField("Average First Response Time (Weekly)", formatNullableTime(firstResponseTime.Weekly), true).
		AddField("Average Ticket Duration (Total)", formatNullableTime(ticketDuration.AllTime), true).
		AddField("Average Ticket Duration (Monthly)", formatNullableTime(ticketDuration.Monthly), true).
		AddField("Average Ticket Duration (Weekly)", formatNullableTime(ticketDuration.Weekly), true)

	_, _ = ctx.ReplyWith(command.NewEphemeralEmbedMessageResponse(msgEmbed))
	span.Finish()
}

func formatNullableTime(duration *time.Duration) string {
	return utils.FormatNullableTime(duration)
}
