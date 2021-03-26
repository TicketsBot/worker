package statistics

import (
	"context"
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"golang.org/x/sync/errgroup"
	"strconv"
	"time"
)

type StatsServerCommand struct {
}

func (StatsServerCommand) Properties() registry.Properties {
	return registry.Properties{
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

func (StatsServerCommand) Execute(ctx registry.CommandContext) {
	var totalTickets, openTickets int

	group, _ := errgroup.WithContext(context.Background())

	// totalTickets
	group.Go(func() (err error) {
		totalTickets, err = dbclient.Client.Tickets.GetTotalTicketCount(ctx.GuildId())
		return
	})

	// openTickets
	group.Go(func() error {
		tickets, err := dbclient.Client.Tickets.GetGuildOpenTickets(ctx.GuildId())
		openTickets = len(tickets)
		return err
	})

	// first response times
	var weekly, monthly, total *time.Duration

	// total
	group.Go(func() (err error) {
		total, err = dbclient.Client.FirstResponseTime.GetAverageAllTime(ctx.GuildId())
		return
	})

	// monthly
	group.Go(func() (err error) {
		monthly, err = dbclient.Client.FirstResponseTime.GetAverage(ctx.GuildId(), time.Hour * 24 * 28)
		return
	})

	// weekly
	group.Go(func() (err error) {
		weekly, err = dbclient.Client.FirstResponseTime.GetAverage(ctx.GuildId(), time.Hour * 24 * 7)
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

	ctx.ReplyWithEmbed(embed)
	ctx.Accept()
}
