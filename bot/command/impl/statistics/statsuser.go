package statistics

import (
	"context"
	"fmt"
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"golang.org/x/sync/errgroup"
	"strconv"
	"time"
)

type StatsUserCommand struct {
}

func (StatsUserCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "user",
		Description:     translations.HelpStats, // TODO: Proper translations
		Aliases:         []string{"statistics"},
		PermissionLevel: permission.Support,
		Category:        command.Statistics,
		PremiumOnly:     true,
		Arguments: command.Arguments(
			command.NewRequiredArgument("user", "User whose statistics to retrieve", interaction.OptionTypeUser, translations.MessageInvalidUser),
		),
	}
}

func (c StatsUserCommand) GetExecutor() interface{} {
	return c.Execute
}

func (StatsUserCommand) Execute(ctx command.CommandContext, userId uint64) {
	member, err := ctx.Worker().GetGuildMember(ctx.GuildId(), userId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	permLevel := permission.GetPermissionLevel(utils.ToRetriever(ctx.Worker()), member, ctx.GuildId())

	// User stats
	if permLevel == 0 {
		var isBlacklisted bool
		var totalTickets int
		var openTickets int
		var ticketLimit uint8

		group, _ := errgroup.WithContext(context.Background())

		// load isBlacklisted
		group.Go(func() (err error) {
			isBlacklisted, err = dbclient.Client.Blacklist.IsBlacklisted(ctx.GuildId(), userId)
			return
		})

		// load totalTickets
		group.Go(func() error {
			tickets, err := dbclient.Client.Tickets.GetAllByUser(ctx.GuildId(), userId)
			totalTickets = len(tickets)
			return err
		})

		// load openTickets
		group.Go(func() error {
			tickets, err := dbclient.Client.Tickets.GetOpenByUser(ctx.GuildId(), userId)
			openTickets = len(tickets)
			return err
		})

		// load ticketLimit
		group.Go(func() (err error) {
			ticketLimit, err = dbclient.Client.TicketLimit.Get(ctx.GuildId())
			return
		})

		if err := group.Wait(); err != nil {
			ctx.HandleError(err)
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

		ctx.ReplyWithEmbed(embed)
	} else { // Support rep stats
		var weekly, monthly, total *time.Duration

		group, _ := errgroup.WithContext(context.Background())

		// total
		group.Go(func() (err error) {
			total, err = dbclient.Client.FirstResponseTime.GetAverageAllTimeUser(ctx.GuildId(), userId)
			return
		})

		// monthly
		group.Go(func() (err error) {
			monthly, err = dbclient.Client.FirstResponseTime.GetAverageUser(ctx.GuildId(), userId, time.Hour*24*28)
			return
		})

		// weekly
		group.Go(func() (err error) {
			weekly, err = dbclient.Client.FirstResponseTime.GetAverageUser(ctx.GuildId(), userId, time.Hour*24*7)
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

			AddField("Is Admin", strconv.FormatBool(permLevel == permission.Admin), true).
			AddField("Is Support", strconv.FormatBool(permLevel >= permission.Support), true).

			AddBlankField(false).

			AddField("Average First Response Time (Total)", totalFormatted, true).
			AddField("Average First Response Time (Monthly)", monthlyFormatted, true).
			AddField("Average First Response Time (Weekly)", weeklyFormatted, true)

		ctx.ReplyWithEmbed(embed)
	}
}
