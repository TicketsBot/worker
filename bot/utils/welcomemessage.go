package utils

import (
	"context"
	"fmt"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/guild/emoji"
	"github.com/rxdn/gdl/objects/interaction/component"
	"github.com/rxdn/gdl/rest"
	"golang.org/x/sync/errgroup"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// returns msg id
func SendWelcomeMessage(ctx registry.CommandContext, ticket database.Ticket, subject string, panel *database.Panel, formData map[database.FormInput]string) (uint64, error) {
	settings, err := dbclient.Client.Settings.Get(ticket.GuildId)
	if err != nil {
		return 0, err
	}

	// Build embeds
	embeds := Slice(BuildWelcomeMessageEmbed(ctx, ticket, subject, panel))

	// Put form fields in a separate embed
	fields := getFormDataFields(formData)
	if len(fields) > 0 {
		formAnswersEmbed := embed.NewEmbed().
			SetColor(ctx.GetColour(customisation.Green))

		for _, field := range fields {
			formAnswersEmbed.AddField(field.Name, field.Value, field.Inline)
		}

		if ctx.PremiumTier() == premium.None {
			formAnswersEmbed.SetFooter("Powered by ticketsbot.net", "https://ticketsbot.net/assets/img/logo.png")
		}

		embeds = append(embeds, formAnswersEmbed)
	}

	buttons := []component.Component{
		component.BuildButton(component.Button{
			Label:    ctx.GetMessage(i18n.TitleClose),
			CustomId: "close",
			Style:    component.ButtonStyleDanger,
			Emoji:    &emoji.Emoji{Name: "🔒"},
		}),
		component.BuildButton(component.Button{
			Label:    ctx.GetMessage(i18n.TitleCloseWithReason),
			CustomId: "close_with_reason",
			Style:    component.ButtonStyleDanger,
			Emoji:    &emoji.Emoji{Name: "🔒"},
		}),
	}

	if !settings.HideClaimButton {
		buttons = append(buttons, component.BuildButton(component.Button{
			Label:    ctx.GetMessage(i18n.TitleClaim),
			CustomId: "claim",
			Style:    component.ButtonStyleSuccess,
			Emoji:    &emoji.Emoji{Name: "🙋‍♂️"},
		}))
	}

	data := rest.CreateMessageData{
		Embeds: embeds,
		Components: []component.Component{
			component.BuildActionRow(buttons...),
		},
	}

	// Should never happen
	if ticket.ChannelId == nil {
		return 0, fmt.Errorf("channel is nil")
	}

	msg, err := ctx.Worker().CreateMessageComplex(*ticket.ChannelId, data)
	if err != nil {
		return 0, err
	}

	return msg.Id, nil
}

func BuildWelcomeMessageEmbed(ctx registry.CommandContext, ticket database.Ticket, subject string, panel *database.Panel) *embed.Embed {
	// Send welcome message
	var welcomeMessage string
	if panel == nil || panel.WelcomeMessage == nil {
		var err error
		welcomeMessage, err = dbclient.Client.WelcomeMessages.Get(ticket.GuildId)
		if err != nil {
			sentry.Error(err)
			welcomeMessage = "Thank you for contacting support.\nPlease describe your issue (and provide an invite to your server if applicable) and wait for a response."
		}
	} else {
		welcomeMessage = *panel.WelcomeMessage
	}

	// %average_response%
	if ctx.PremiumTier() > premium.None && strings.Contains(welcomeMessage, "%average_response%") {
		weeklyResponseTime, err := dbclient.Client.FirstResponseTime.GetAverage(ticket.GuildId, time.Hour*24*7)
		if err != nil {
			sentry.Error(err)
		} else {
			strings.Replace(welcomeMessage, "%average_response%", FormatTime(*weeklyResponseTime), -1)
		}
	}

	// variables
	welcomeMessage = DoPlaceholderSubstitutions(welcomeMessage, ctx.Worker(), ticket)

	return BuildEmbedRaw(ctx.GetColour(customisation.Green), subject, welcomeMessage, nil, ctx.PremiumTier())
}

func DoPlaceholderSubstitutions(message string, ctx *worker.Context, ticket database.Ticket) string {
	var lock sync.Mutex

	// do DB lookups in parallel
	group, _ := errgroup.WithContext(context.Background())
	for placeholder, f := range substitutions {
		placeholder := placeholder
		f := f

		formatted := fmt.Sprintf("%%%s%%", placeholder)

		if strings.Contains(message, formatted) {
			group.Go(func() error {
				replacement := f(ctx, ticket)

				lock.Lock()
				message = strings.Replace(message, formatted, replacement, -1)
				lock.Unlock()

				return nil
			})
		}
	}

	if err := group.Wait(); err != nil {
		sentry.Error(err)
	}

	return message
}

var substitutions = map[string]func(ctx *worker.Context, ticket database.Ticket) string{
	"user": func(ctx *worker.Context, ticket database.Ticket) string {
		return fmt.Sprintf("<@%d>", ticket.UserId)
	},
	"ticket_id": func(ctx *worker.Context, ticket database.Ticket) string {
		return strconv.Itoa(ticket.Id)
	},
	"channel": func(ctx *worker.Context, ticket database.Ticket) string {
		return fmt.Sprintf("<#%d>", ticket.ChannelId)
	},
	"username": func(ctx *worker.Context, ticket database.Ticket) string {
		user, _ := ctx.GetUser(ticket.UserId)
		return user.Username
	},
	"server": func(ctx *worker.Context, ticket database.Ticket) string {
		guild, _ := ctx.GetGuild(ticket.GuildId)
		return guild.Name
	},
	"open_tickets": func(ctx *worker.Context, ticket database.Ticket) string {
		open, _ := dbclient.Client.Tickets.GetGuildOpenTickets(ticket.GuildId)
		return strconv.Itoa(len(open))
	},
	"total_tickets": func(ctx *worker.Context, ticket database.Ticket) string {
		total, _ := dbclient.Client.Tickets.GetTotalTicketCount(ticket.GuildId)
		return strconv.Itoa(total)
	},
	"user_open_tickets": func(ctx *worker.Context, ticket database.Ticket) string {
		tickets, _ := dbclient.Client.Tickets.GetOpenByUser(ticket.GuildId, ticket.UserId)
		return strconv.Itoa(len(tickets))
	},
	"ticket_limit": func(ctx *worker.Context, ticket database.Ticket) string {
		limit, _ := dbclient.Client.TicketLimit.Get(ticket.GuildId)
		return strconv.Itoa(int(limit))
	},
	"rating_count": func(ctx *worker.Context, ticket database.Ticket) string {
		ratingCount, _ := dbclient.Client.ServiceRatings.GetCount(ticket.GuildId)
		return strconv.Itoa(ratingCount)
	},
	"average_rating": func(ctx *worker.Context, ticket database.Ticket) string {
		average, _ := dbclient.Client.ServiceRatings.GetAverage(ticket.GuildId)
		return fmt.Sprintf("%.1f", average)
	},
	"time": func(ctx *worker.Context, ticket database.Ticket) string {
		return fmt.Sprintf("<t:%d:t>", time.Now().Unix())
	},
	"date": func(ctx *worker.Context, ticket database.Ticket) string {
		return fmt.Sprintf("<t:%d:d>", time.Now().Unix())
	},
	"datetime": func(ctx *worker.Context, ticket database.Ticket) string {
		return fmt.Sprintf("<t:%d:f>", time.Now().Unix())
	},
	"first_response_time_weekly": func(ctx *worker.Context, ticket database.Ticket) string {
		data, _ := dbclient.Client.FirstResponseTimeGuildView.Get(ticket.GuildId)
		return FormatNullableTime(data.Weekly)
	},
	"first_response_time_monthly": func(ctx *worker.Context, ticket database.Ticket) string {
		data, _ := dbclient.Client.FirstResponseTimeGuildView.Get(ticket.GuildId)
		return FormatNullableTime(data.Monthly)
	},
	"first_response_time_all_time": func(ctx *worker.Context, ticket database.Ticket) string {
		data, _ := dbclient.Client.FirstResponseTimeGuildView.Get(ticket.GuildId)
		return FormatNullableTime(data.AllTime)
	},
}

func getFormDataFields(formData map[database.FormInput]string) []embed.EmbedField {
	// Get form inputs in the same order they are presented on the dashboard
	i := 0
	inputs := make([]database.FormInput, len(formData))
	for input := range formData {
		inputs[i] = input
		i++
	}

	sort.Slice(inputs, func(i, j int) bool {
		return inputs[i].Position < inputs[j].Position
	})

	var fields []embed.EmbedField // Can't use len(formData), as form may have changed since modal was opened
	for _, input := range inputs {
		answer, ok := formData[input]
		if answer == "" {
			answer = "N/A" // TODO: What should we use here?
		}

		if ok {
			fields = append(fields, embed.EmbedField{
				Name:   input.Label,
				Value:  answer,
				Inline: false,
			})
		}
	}

	return fields
}
