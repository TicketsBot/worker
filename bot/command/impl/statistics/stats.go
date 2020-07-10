package statistics

import (
	"context"
	"fmt"
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

type StatsCommand struct {
}

func (StatsCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "stats",
		Description:     translations.HelpStats,
		Aliases:         []string{"statistics"},
		PermissionLevel: permission.Support,
		Children: []command.Command{
			StatsServerCommand{},
		},
		Category:    command.Statistics,
		PremiumOnly: true,
	}
}

func (StatsCommand) Execute(ctx command.CommandContext) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!stats server`\n`t!stats @User`",
		Inline: false,
	}

	if len(ctx.Args) == 0 {
		ctx.SendEmbed(utils.Red, "Error", translations.MessageInvalidArgument, usageEmbed)
		ctx.ReactWithCross()
		return
	}

	// server is handled as a subcommand, so a user has been pinged
	if len(ctx.Message.Mentions) == 0 {
		ctx.SendEmbed(utils.Red, "Error", translations.MessageInvalidArgument, usageEmbed)
		ctx.ReactWithCross()
		return
	}

	user := ctx.Message.Mentions[0]

	// User stats
	if ctx.UserPermissionLevel == 0 {
		var isBlacklisted bool
		var totalTickets int
		var openTickets int
		var ticketLimit uint8

		group, _ := errgroup.WithContext(context.Background())

		// load isBlacklisted
		group.Go(func() (err error) {
			isBlacklisted, err = dbclient.Client.Blacklist.IsBlacklisted(ctx.GuildId, user.Id)
			return
		})

		// load totalTickets
		group.Go(func() error {
			tickets, err := dbclient.Client.Tickets.GetAllByUser(ctx.GuildId, user.Id)
			totalTickets = len(tickets)
			return err
		})

		// load openTickets
		group.Go(func() error {
			tickets, err := dbclient.Client.Tickets.GetOpenByUser(ctx.GuildId, user.Id)
			openTickets = len(tickets)
			return err
		})

		// load ticketLimit
		group.Go(func() (err error) {
			ticketLimit, err = dbclient.Client.TicketLimit.Get(ctx.GuildId)
			return
		})

		if err := group.Wait(); err != nil {
			ctx.ReactWithCross()
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			return
		}

		embed := embed.NewEmbed().
			SetTitle("Statistics").
			SetColor(int(utils.Green)).

			AddField("Is Admin", "false", true).
			AddField("Is Support", "false", true).
			AddField("Is Blacklisted", strconv.FormatBool(isBlacklisted), true).

			AddField("Total Tickets", strconv.Itoa(totalTickets), true).
			AddField("Open Tickets", fmt.Sprintf("%d / %d", openTickets, ticketLimit), true)

		if m, err := ctx.Worker.CreateMessageEmbed(ctx.ChannelId, embed); err == nil {
			utils.DeleteAfter(utils.SentMessage{Worker: ctx.Worker, Message: &m}, 60)
		}
	} else { // Support rep stats
		var weekly, monthly, total *time.Duration

		group, _ := errgroup.WithContext(context.Background())

		// total
		group.Go(func() (err error) {
			total, err = dbclient.Client.FirstResponseTime.GetAverageAllTimeUser(ctx.GuildId, user.Id)
			return
		})

		// monthly
		group.Go(func() (err error) {
			monthly, err = dbclient.Client.FirstResponseTime.GetAverageUser(ctx.GuildId, user.Id, time.Hour*24*28)
			return
		})

		// weekly
		group.Go(func() (err error) {
			weekly, err = dbclient.Client.FirstResponseTime.GetAverageUser(ctx.GuildId, user.Id, time.Hour*24*7)
			return
		})

		if err := group.Wait(); err != nil {
			ctx.ReactWithCross()
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
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

			AddField("Is Admin", strconv.FormatBool(ctx.UserPermissionLevel == permission.Admin), true).
			AddField("Is Support", strconv.FormatBool(ctx.UserPermissionLevel >= permission.Support), true).

			AddBlankField(false).

			AddField("Average First Response Time (Total)", totalFormatted, true).
			AddField("Average First Response Time (Monthly)", monthlyFormatted, true).
			AddField("Average First Response Time (Weekly)", weeklyFormatted, true)

		if m, err := ctx.Worker.CreateMessageEmbed(ctx.ChannelId, embed); err == nil {
			utils.DeleteAfter(utils.SentMessage{Worker: ctx.Worker, Message: &m}, 60)
		}
	}
}
