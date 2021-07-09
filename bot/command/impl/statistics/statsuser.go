package statistics

import (
	"context"
	"fmt"
	"github.com/TicketsBot/common/permission"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
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

func (StatsUserCommand) Properties() registry.Properties {
	return registry.Properties{
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

func (StatsUserCommand) Execute(ctx registry.CommandContext, userId uint64) {
	member, err := ctx.Worker().GetGuildMember(ctx.GuildId(), userId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	permLevel, err := permission.GetPermissionLevel(utils.ToRetriever(ctx.Worker()), member, ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// User stats
	if permLevel == permission.Everyone {
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
		var weeklyAR, monthlyAR, totalAR *time.Duration
		var weeklyAnsweredTickets, monthlyAnsweredTickets, totalAnsweredTickets,
		weeklyTotalTickets, monthlyTotalTickets, totalTotalTickets,
		weeklyClaimedTickets, monthlyClaimedTickets, totalClaimedTickets int

		group, _ := errgroup.WithContext(context.Background())

		// totalAR
		group.Go(func() (err error) {
			totalAR, err = dbclient.Client.FirstResponseTime.GetAverageAllTimeUser(ctx.GuildId(), userId)
			return
		})

		// monthlyAR
		group.Go(func() (err error) {
			monthlyAR, err = dbclient.Client.FirstResponseTime.GetAverageUser(ctx.GuildId(), userId, time.Hour*24*28)
			return
		})

		// weeklyAR
		group.Go(func() (err error) {
			weeklyAR, err = dbclient.Client.FirstResponseTime.GetAverageUser(ctx.GuildId(), userId, time.Hour*24*7)
			return
		})

		// weeklyAnswered
		group.Go(func() (err error) {
			weeklyAnsweredTickets, err = dbclient.Client.Participants.GetParticipatedCountInterval(ctx.GuildId(), userId, time.Hour*24*7)
			return
		})

		// monthlyAnswered
		group.Go(func() (err error) {
			monthlyAnsweredTickets, err = dbclient.Client.Participants.GetParticipatedCountInterval(ctx.GuildId(), userId, time.Hour*24*28)
			return
		})

		// totalAnswered
		group.Go(func() (err error) {
			totalAnsweredTickets, err = dbclient.Client.Participants.GetParticipatedCount(ctx.GuildId(), userId)
			return
		})

		// weeklyTotal
		group.Go(func() (err error) {
			weeklyTotalTickets, err = dbclient.Client.Tickets.GetTotalTicketCountInterval(ctx.GuildId(), time.Hour*24*7)
			return
		})

		// monthlyTotal
		group.Go(func() (err error) {
			monthlyTotalTickets, err = dbclient.Client.Tickets.GetTotalTicketCountInterval(ctx.GuildId(), time.Hour*24*28)
			return
		})

		// totalTotal
		group.Go(func() (err error) {
			totalTotalTickets, err = dbclient.Client.Tickets.GetTotalTicketCount(ctx.GuildId())
			return
		})

		// weeklyClaimed
		group.Go(func() (err error) {
			weeklyClaimedTickets, err = dbclient.Client.TicketClaims.GetClaimedSinceCount(ctx.GuildId(), userId, time.Hour*24*7)
			return
		})

		// monthlyClaimed
		group.Go(func() (err error) {
			monthlyClaimedTickets, err = dbclient.Client.TicketClaims.GetClaimedSinceCount(ctx.GuildId(), userId, time.Hour*24*28)
			return
		})

		// totalClaimed
		group.Go(func() (err error) {
			totalClaimedTickets, err = dbclient.Client.TicketClaims.GetClaimedCount(ctx.GuildId(), userId)
			return
		})

		if err := group.Wait(); err != nil {
			ctx.HandleError(err)
			return
		}

		var totalFormatted, monthlyFormatted, weeklyFormatted string

		if totalAR == nil {
			totalFormatted = "No data"
		} else {
			totalFormatted = utils.FormatTime(*totalAR)
		}

		if monthlyAR == nil {
			monthlyFormatted = "No data"
		} else {
			monthlyFormatted = utils.FormatTime(*monthlyAR)
		}

		if weeklyAR == nil {
			weeklyFormatted = "No data"
		} else {
			weeklyFormatted = utils.FormatTime(*weeklyAR)
		}

		embed := embed.NewEmbed().
			SetTitle("Statistics").
			SetColor(int(utils.Green)).

			AddField("Is Admin", strconv.FormatBool(permLevel == permission.Admin), true).
			AddField("Is Support", strconv.FormatBool(permLevel >= permission.Support), true).
			AddBlankField(true).

			AddField("Average First Response Time (Weekly)", weeklyFormatted, true).
			AddField("Average First Response Time (Monthly)", monthlyFormatted, true).
			AddField("Average First Response Time (Total)", totalFormatted, true).

			AddField("Tickets Answered (Weekly)", fmt.Sprintf("%d / %d", weeklyAnsweredTickets, weeklyTotalTickets), true).
			AddField("Tickets Answered (Monthly)", fmt.Sprintf("%d / %d", monthlyAnsweredTickets, monthlyTotalTickets), true).
			AddField("Tickets Answered (Total)", fmt.Sprintf("%d / %d", totalAnsweredTickets, totalTotalTickets), true).

			AddField("Claimed Tickets (Weekly)", strconv.Itoa(weeklyClaimedTickets), true).
			AddField("Claimed Tickets (Monthly)", strconv.Itoa(monthlyClaimedTickets), true).
			AddField("Claimed Tickets (Total)", strconv.Itoa(totalClaimedTickets), true)

		ctx.ReplyWithEmbed(embed)
	}
}
