package statistics

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/getsentry/sentry-go"
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
		Description:     i18n.HelpStats,
		Type:            interaction.ApplicationCommandTypeChatInput,
		Aliases:         []string{"statistics"},
		PermissionLevel: permission.Support,
		Category:        command.Statistics,
		PremiumOnly:     true,
		Arguments: command.Arguments(
			command.NewRequiredArgument("user", "User whose statistics to retrieve", interaction.OptionTypeUser, i18n.MessageInvalidUser),
		),
		DefaultEphemeral: true,
		Timeout:          time.Second * 30,
	}
}

func (c StatsUserCommand) GetExecutor() interface{} {
	return c.Execute
}

func (StatsUserCommand) Execute(ctx registry.CommandContext, userId uint64) {
	span := sentry.StartTransaction(ctx, "/stats user")
	span.SetTag("guild", strconv.FormatUint(ctx.GuildId(), 10))
	span.SetTag("user", strconv.FormatUint(userId, 10))
	defer span.Finish()

	member, err := ctx.Worker().GetGuildMember(ctx.GuildId(), userId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	permLevel, err := permission.GetPermissionLevel(ctx, utils.ToRetriever(ctx.Worker()), member, ctx.GuildId())
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

		group, _ := errgroup.WithContext(ctx)

		// load isBlacklisted
		group.Go(func() (err error) {
			span := sentry.StartSpan(span.Context(), "Is Blacklisted")
			defer span.Finish()

			isBlacklisted, err = utils.IsBlacklisted(ctx, ctx.GuildId(), userId, member, permLevel)
			return
		})

		// load totalTickets
		group.Go(func() error {
			span := sentry.StartSpan(span.Context(), "GetAllByUser")
			defer span.Finish()

			tickets, err := dbclient.Client.Tickets.GetAllByUser(ctx, ctx.GuildId(), userId)
			totalTickets = len(tickets)
			return err
		})

		// load openTickets
		group.Go(func() error {
			span := sentry.StartSpan(span.Context(), "GetOpenByUser")
			defer span.Finish()

			tickets, err := dbclient.Client.Tickets.GetOpenByUser(ctx, ctx.GuildId(), userId)
			openTickets = len(tickets)
			return err
		})

		// load ticketLimit
		group.Go(func() (err error) {
			span := sentry.StartSpan(span.Context(), "TicketLimit")
			defer span.Finish()

			ticketLimit, err = dbclient.Client.TicketLimit.Get(ctx, ctx.GuildId())
			return
		})

		if err := group.Wait(); err != nil {
			ctx.HandleError(err)
			return
		}

		span := sentry.StartSpan(span.Context(), "Reply")

		msgEmbed := embed.NewEmbed().
			SetTitle("Statistics").
			SetColor(ctx.GetColour(customisation.Green)).
			SetAuthor(member.User.Username, "", member.User.AvatarUrl(256)).
			AddField("Permission Level", "Regular", true).
			AddField("Is Blacklisted", strconv.FormatBool(isBlacklisted), true).
			AddBlankField(true).
			AddField("Total Tickets", strconv.Itoa(totalTickets), true).
			AddField("Open Tickets", fmt.Sprintf("%d / %d", openTickets, ticketLimit), true)

		_, _ = ctx.ReplyWith(command.NewEphemeralEmbedMessageResponse(msgEmbed))
		span.Finish()
	} else { // Support rep stats
		group, _ := errgroup.WithContext(ctx)

		var feedbackRating float32
		var feedbackCount int

		group.Go(func() (err error) {
			span := sentry.StartSpan(span.Context(), "GetAverageClaimedBy")
			defer span.Finish()

			feedbackRating, err = dbclient.Client.ServiceRatings.GetAverageClaimedBy(ctx, ctx.GuildId(), userId)
			return
		})

		group.Go(func() (err error) {
			span := sentry.StartSpan(span.Context(), "GetCountClaimedBy")
			defer span.Finish()

			feedbackCount, err = dbclient.Client.ServiceRatings.GetCountClaimedBy(ctx, ctx.GuildId(), userId)
			return
		})

		var weeklyAR, monthlyAR, totalAR *time.Duration
		var weeklyAnsweredTickets, monthlyAnsweredTickets, totalAnsweredTickets,
			weeklyTotalTickets, monthlyTotalTickets, totalTotalTickets,
			weeklyClaimedTickets, monthlyClaimedTickets, totalClaimedTickets int

		// totalAR
		group.Go(func() (err error) {
			span := sentry.StartSpan(span.Context(), "GetAverageAllTimeUser")
			defer span.Finish()

			totalAR, err = dbclient.Client.FirstResponseTime.GetAverageAllTimeUser(ctx, ctx.GuildId(), userId)
			return
		})

		// monthlyAR
		group.Go(func() (err error) {
			span := sentry.StartSpan(span.Context(), "GetAverageUser")
			defer span.Finish()

			monthlyAR, err = dbclient.Client.FirstResponseTime.GetAverageUser(ctx, ctx.GuildId(), userId, time.Hour*24*28)
			return
		})

		// weeklyAR
		group.Go(func() (err error) {
			span := sentry.StartSpan(span.Context(), "GetAverageUser")
			defer span.Finish()

			weeklyAR, err = dbclient.Client.FirstResponseTime.GetAverageUser(ctx, ctx.GuildId(), userId, time.Hour*24*7)
			return
		})

		// weeklyAnswered
		group.Go(func() (err error) {
			span := sentry.StartSpan(span.Context(), "GetParticipatedCountInterval")
			defer span.Finish()

			weeklyAnsweredTickets, err = dbclient.Client.Participants.GetParticipatedCountInterval(ctx, ctx.GuildId(), userId, time.Hour*24*7)
			return
		})

		// monthlyAnswered
		group.Go(func() (err error) {
			span := sentry.StartSpan(span.Context(), "GetParticipatedCountInterval")
			defer span.Finish()

			monthlyAnsweredTickets, err = dbclient.Client.Participants.GetParticipatedCountInterval(ctx, ctx.GuildId(), userId, time.Hour*24*28)
			return
		})

		// totalAnswered
		group.Go(func() (err error) {
			span := sentry.StartSpan(span.Context(), "GetParticipatedCount")
			defer span.Finish()

			totalAnsweredTickets, err = dbclient.Client.Participants.GetParticipatedCount(ctx, ctx.GuildId(), userId)
			return
		})

		// weeklyTotal
		group.Go(func() (err error) {
			span := sentry.StartSpan(span.Context(), "GetTotalTicketCountInterval")
			defer span.Finish()

			weeklyTotalTickets, err = dbclient.Client.Tickets.GetTotalTicketCountInterval(ctx, ctx.GuildId(), time.Hour*24*7)
			return
		})

		// monthlyTotal
		group.Go(func() (err error) {
			span := sentry.StartSpan(span.Context(), "GetTotalTicketCountInterval")
			defer span.Finish()

			monthlyTotalTickets, err = dbclient.Client.Tickets.GetTotalTicketCountInterval(ctx, ctx.GuildId(), time.Hour*24*28)
			return
		})

		// totalTotal
		group.Go(func() (err error) {
			span := sentry.StartSpan(span.Context(), "GetTotalTicketCount")
			defer span.Finish()

			totalTotalTickets, err = dbclient.Client.Tickets.GetTotalTicketCount(ctx, ctx.GuildId())
			return
		})

		// weeklyClaimed
		group.Go(func() (err error) {
			span := sentry.StartSpan(span.Context(), "GetClaimedSinceCount_Weekly")
			defer span.Finish()

			weeklyClaimedTickets, err = dbclient.Client.TicketClaims.GetClaimedSinceCount(ctx, ctx.GuildId(), userId, time.Hour*24*7)
			return
		})

		// monthlyClaimed
		group.Go(func() (err error) {
			span := sentry.StartSpan(span.Context(), "GetClaimedSinceCount_Monthly")
			defer span.Finish()

			monthlyClaimedTickets, err = dbclient.Client.TicketClaims.GetClaimedSinceCount(ctx, ctx.GuildId(), userId, time.Hour*24*28)
			return
		})

		// totalClaimed
		group.Go(func() (err error) {
			span := sentry.StartSpan(span.Context(), "GetClaimedCount")
			defer span.Finish()

			totalClaimedTickets, err = dbclient.Client.TicketClaims.GetClaimedCount(ctx, ctx.GuildId(), userId)
			return
		})

		if err := group.Wait(); err != nil {
			ctx.HandleError(err)
			return
		}

		var permissionLevel string
		if permLevel == permission.Admin {
			permissionLevel = "Admin"
		} else {
			permissionLevel = "Support"
		}

		span := sentry.StartSpan(span.Context(), "Reply")

		msgEmbed := embed.NewEmbed().
			SetTitle("Statistics").
			SetColor(ctx.GetColour(customisation.Green)).
			SetAuthor(member.User.Username, "", member.User.AvatarUrl(256)).
			AddField("Permission Level", permissionLevel, true).
			AddField("Feedback Rating", fmt.Sprintf("%.1f / 5 ⭐ (%d ratings)", feedbackRating, feedbackCount), true).
			AddBlankField(true).
			AddField("Average First Response Time (Weekly)", formatNullableTime(weeklyAR), true).
			AddField("Average First Response Time (Monthly)", formatNullableTime(monthlyAR), true).
			AddField("Average First Response Time (Total)", formatNullableTime(totalAR), true).
			AddField("Tickets Answered (Weekly)", fmt.Sprintf("%d / %d", weeklyAnsweredTickets, weeklyTotalTickets), true).
			AddField("Tickets Answered (Monthly)", fmt.Sprintf("%d / %d", monthlyAnsweredTickets, monthlyTotalTickets), true).
			AddField("Tickets Answered (Total)", fmt.Sprintf("%d / %d", totalAnsweredTickets, totalTotalTickets), true).
			AddField("Claimed Tickets (Weekly)", strconv.Itoa(weeklyClaimedTickets), true).
			AddField("Claimed Tickets (Monthly)", strconv.Itoa(monthlyClaimedTickets), true).
			AddField("Claimed Tickets (Total)", strconv.Itoa(totalClaimedTickets), true)

		_, _ = ctx.ReplyWith(command.NewEphemeralEmbedMessageResponse(msgEmbed))
		span.Finish()
	}
}
