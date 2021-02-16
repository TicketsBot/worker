package statistics

import (
	"context"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"golang.org/x/sync/errgroup"
	"strconv"
	"time"
)

type StatsServerCommand struct {
}

func (StatsServerCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "server",
		Description:     translations.HelpStatsServer,
		PermissionLevel: permission.Support,
		Category:        command.Statistics,
		PremiumOnly:     true,
	}
}

func (c StatsServerCommand) GetExecutor() interface{} {
	return c.Execute
}

func (StatsServerCommand) Execute(ctx command.CommandContext) {
	var totalTickets, openTickets int

	tags := map[string]string{
		"guild": strconv.FormatUint(ctx.GuildId(), 10),
	}

	sentry.LogWithTags("stats 1", nil, tags)

	group, _ := errgroup.WithContext(context.Background())

	// totalTickets
	group.Go(func() (err error) {
		totalTickets, err = dbclient.Client.Tickets.GetTotalTicketCount(ctx.GuildId())
		sentry.LogWithTags("stats 2", map[string]interface{}{"total_tickets": totalTickets}, tags)
		return
	})

	// openTickets
	group.Go(func() error {
		tickets, err := dbclient.Client.Tickets.GetGuildOpenTickets(ctx.GuildId())
		openTickets = len(tickets)
		sentry.LogWithTags("stats 3", map[string]interface{}{"open_tickets": openTickets}, tags)
		return err
	})

	// first response times
	var weekly, monthly, total *time.Duration

	// total
	group.Go(func() (err error) {
		total, err = dbclient.Client.FirstResponseTime.GetAverageAllTime(ctx.GuildId())
		sentry.LogWithTags("stats 4", map[string]interface{}{"first_response_total": total}, tags)
		return
	})

	// monthly
	group.Go(func() (err error) {
		monthly, err = dbclient.Client.FirstResponseTime.GetAverage(ctx.GuildId(), time.Hour * 24 * 28)
		sentry.LogWithTags("stats 5", map[string]interface{}{"first_response_monthly": monthly}, tags)
		return
	})

	// weekly
	group.Go(func() (err error) {
		weekly, err = dbclient.Client.FirstResponseTime.GetAverage(ctx.GuildId(), time.Hour * 24 * 7)
		sentry.LogWithTags("stats 6", map[string]interface{}{"first_response_weekly": weekly}, tags)
		return
	})

	if err := group.Wait(); err != nil {
		ctx.HandleError(err)
		return
	}

	var totalFormatted, monthlyFormatted, weeklyFormatted string

	if total == nil {
		totalFormatted = "No data"
	} else {
		totalFormatted = utils.FormatTime(*total)
	}

	if monthly == nil {
		monthlyFormatted = "No data"
	} else {
		monthlyFormatted = utils.FormatTime(*monthly)
	}

	if weekly == nil {
		weeklyFormatted = "No data"
	} else {
		weeklyFormatted = utils.FormatTime(*weekly)
	}

	embed := embed.NewEmbed().
		SetTitle("Statistics").
		SetColor(int(utils.Green)).

		AddField("Total Tickets", strconv.Itoa(totalTickets), true).
		AddField("Open Tickets", strconv.Itoa(openTickets), true).

		AddBlankField(false).

		AddField("Average First Response Time (Total)", totalFormatted, true).
		AddField("Average First Response Time (Monthly)", monthlyFormatted, true).
		AddField("Average First Response Time (Weekly)", weeklyFormatted, true)

	sentry.LogWithTags("stats 7", nil, tags)
	ctx.ReplyWithEmbed(embed)
	ctx.Accept()
}
