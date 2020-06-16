package utils

import (
	"fmt"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"strconv"
	"strings"
	"time"
)

// returns msg id
func SendWelcomeMessage(worker *worker.Context, guildId, channelId, userId uint64, isPremium bool, subject string, panel *database.Panel, ticketId int) (uint64, error) {
	// Send welcome message
	var welcomeMessage string
	if panel == nil || panel.WelcomeMessage == nil {
		var err error
		welcomeMessage, err = dbclient.Client.WelcomeMessages.Get(guildId); if err != nil {
			sentry.Error(err)
			welcomeMessage = "Thank you for contacting support.\nPlease describe your issue (and provide an invite to your server if applicable) and wait for a response."
		}
	} else {
		welcomeMessage = *panel.WelcomeMessage
	}

	// %average_response%
	if isPremium && strings.Contains(welcomeMessage, "%average_response%") {
		weeklyResponseTime, err := dbclient.Client.FirstResponseTime.GetAverage(guildId, time.Hour * 24 * 7)
		if err != nil {
			sentry.Error(err)
		} else {
			strings.Replace(welcomeMessage, "%average_response%", FormatTime(*weeklyResponseTime), -1)
		}
	}

	// variables
	welcomeMessage = doSubstitutions(welcomeMessage, worker, guildId, userId, channelId, ticketId)

	// Send welcome message
	if msg, err := SendEmbedWithResponse(worker, channelId, Green, subject, welcomeMessage, nil, 0, isPremium); err == nil {
		// Add close reaction to the welcome message
		err := worker.CreateReaction(channelId, msg.Id, "🔒")
		if err != nil {
			sentry.Error(err)
		}

		return msg.Id, err
	} else {
		return 0, err
	}
}

func doSubstitutions(welcomeMessage string, worker *worker.Context, guildId, userId, channelId uint64, ticketId int) string {
	// %user%
	welcomeMessage = strings.Replace(welcomeMessage, "%user%", fmt.Sprintf("<@%d>", userId), -1)

	// %username%
	if strings.Contains(welcomeMessage, "%username%") {
		if user, err := worker.GetUser(userId); err == nil {
			welcomeMessage = strings.Replace(welcomeMessage, "%username%", user.Username, -1)
		}
	}

	// %server%
	if strings.Contains(welcomeMessage, "%server%") {
		if guild, err := worker.GetGuild(guildId); err == nil {
			welcomeMessage = strings.Replace(welcomeMessage, "%server%", guild.Name, -1)
		}
	}

	// %ticket_id%
	if ticketId > 0 {
		welcomeMessage = strings.Replace(welcomeMessage, "%ticket_id%", strconv.Itoa(ticketId), -1)
	}

	// %open_tickets%
	if strings.Contains(welcomeMessage, "%open_tickets%") {
		if open, err := dbclient.Client.Tickets.GetGuildOpenTickets(guildId); err == nil {
			welcomeMessage = strings.Replace(welcomeMessage, "%open_tickets%", strconv.Itoa(len(open)), -1)
		}
	}

	// %total_tickets%
	if strings.Contains(welcomeMessage, "%total_tickets%") {
		if count, err := dbclient.Client.Tickets.GetTotalTicketCount(guildId); err == nil {
			welcomeMessage = strings.Replace(welcomeMessage, "%total_tickets%", strconv.Itoa(count), -1)
		}
	}

	// %user_open_tickets%
	if strings.Contains(welcomeMessage, "%user_open_tickets%") {
		if open, err := dbclient.Client.Tickets.GetOpenByUser(guildId, userId); err == nil {
			welcomeMessage = strings.Replace(welcomeMessage, "%user_open_tickets%", strconv.Itoa(len(open)), -1)
		}
	}

	// %ticket_limit%
	if strings.Contains(welcomeMessage, "%ticket_limit%") {
		if limit, err := dbclient.Client.TicketLimit.Get(guildId); err == nil {
			welcomeMessage = strings.Replace(welcomeMessage, "%ticket_limit%", strconv.Itoa(int(limit)), -1)
		}
	}

	// %channel%
	welcomeMessage = strings.Replace(welcomeMessage, "%channel%", fmt.Sprintf("<#%d>", channelId), -1)

	return welcomeMessage
}
