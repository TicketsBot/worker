package statistics

import (
	"fmt"
	"github.com/TicketsBot/analytics-client"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/getsentry/sentry-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
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
		Timeout:          time.Second * 10,
	}
}

func (c StatsServerCommand) GetExecutor() interface{} {
	return c.Execute
}

func (StatsServerCommand) Execute(ctx registry.CommandContext) {
	span := sentry.StartTransaction(ctx, "/stats server")
	span.SetTag("guild", strconv.FormatUint(ctx.GuildId(), 10))
	defer span.Finish()

	group, _ := errgroup.WithContext(ctx)

	var totalTickets, openTickets uint64

	// totalTickets
	group.Go(func() (err error) {
		span := sentry.StartSpan(span.Context(), "GetTotalTicketCount")
		defer span.Finish()

		totalTickets, err = dbclient.Analytics.GetTotalTicketCount(ctx, ctx.GuildId())
		return
	})

	// openTickets
	group.Go(func() error {
		span := sentry.StartSpan(span.Context(), "GetGuildOpenTickets")
		defer span.Finish()

		tickets, err := dbclient.Client.Tickets.GetGuildOpenTickets(ctx, ctx.GuildId())
		if err != nil {
			return err
		}

		openTickets = uint64(len(tickets))
		return nil
	})

	var feedbackRating float64
	var feedbackCount uint64

	group.Go(func() (err error) {
		span := sentry.StartSpan(span.Context(), "GetAverageFeedbackRating")
		defer span.Finish()

		feedbackRating, err = dbclient.Analytics.GetAverageFeedbackRatingGuild(ctx, ctx.GuildId())
		return
	})

	group.Go(func() (err error) {
		span := sentry.StartSpan(span.Context(), "GetFeedbackCount")
		defer span.Finish()

		feedbackCount, err = dbclient.Analytics.GetFeedbackCountGuild(ctx, ctx.GuildId())
		return
	})

	// first response times
	var firstResponseTime analytics.TripleWindow
	group.Go(func() (err error) {
		span := sentry.StartSpan(span.Context(), "GetFirstResponseTimeStats")
		defer span.Finish()

		firstResponseTime, err = dbclient.Analytics.GetFirstResponseTimeStats(ctx, ctx.GuildId())
		return
	})

	// ticket duration
	var ticketDuration analytics.TripleWindow
	group.Go(func() (err error) {
		span := sentry.StartSpan(span.Context(), "GetTicketDurationStats")
		defer span.Finish()

		ticketDuration, err = dbclient.Analytics.GetTicketDurationStats(ctx, ctx.GuildId())
		return
	})

	// tickets per day
	var ticketVolumeTable string
	group.Go(func() error {
		span := sentry.StartSpan(span.Context(), "GetLastNTicketsPerDayGuild")
		defer span.Finish()

		counts, err := dbclient.Analytics.GetLastNTicketsPerDayGuild(ctx, ctx.GuildId(), 7)
		if err != nil {
			return err
		}

		tw := table.NewWriter()
		tw.SetStyle(table.StyleLight)
		tw.Style().Format.Header = text.FormatDefault

		tw.AppendHeader(table.Row{"Date", "Ticket Volume"})
		for _, count := range counts {
			tw.AppendRow(table.Row{count.Date.Format("2006-01-02"), count.Count})
		}

		ticketVolumeTable = tw.Render()
		return nil
	})

	if err := group.Wait(); err != nil {
		ctx.HandleError(err)
		return
	}

	span = sentry.StartSpan(span.Context(), "Send Message")

	msgEmbed := embed.NewEmbed().
		SetTitle("Statistics").
		SetColor(ctx.GetColour(customisation.Green)).
		AddField("Total Tickets", strconv.FormatUint(totalTickets, 10), true).
		AddField("Open Tickets", strconv.FormatUint(openTickets, 10), true).
		AddBlankField(true).
		AddField("Feedback Rating", fmt.Sprintf("%.1f / 5 ⭐", feedbackRating), true).
		AddField("Feedback Count", strconv.FormatUint(feedbackCount, 10), true).
		AddBlankField(true).
		AddField("Average First Response Time (Total)", formatNullableTime(firstResponseTime.AllTime), true).
		AddField("Average First Response Time (Monthly)", formatNullableTime(firstResponseTime.Monthly), true).
		AddField("Average First Response Time (Weekly)", formatNullableTime(firstResponseTime.Weekly), true).
		AddField("Average Ticket Duration (Total)", formatNullableTime(ticketDuration.AllTime), true).
		AddField("Average Ticket Duration (Monthly)", formatNullableTime(ticketDuration.Monthly), true).
		AddField("Average Ticket Duration (Weekly)", formatNullableTime(ticketDuration.Weekly), true).
		AddField("Ticket Volume", fmt.Sprintf("```\n%s\n```", ticketVolumeTable), false)

	_, _ = ctx.ReplyWith(command.NewEphemeralEmbedMessageResponse(msgEmbed))
	span.Finish()
}

func formatNullableTime(duration *time.Duration) string {
	return utils.FormatNullableTime(duration)
}
