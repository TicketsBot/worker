package utils

import (
	"context"
	"fmt"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/guild/emoji"
	"github.com/rxdn/gdl/objects/interaction/component"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/rest"
	"golang.org/x/sync/errgroup"
	"strconv"
	"strings"
	"sync"
	"time"
)

// returns msg id
func SendWelcomeMessage(worker *worker.Context, guildId, channelId, userId uint64, isPremium bool, subject string, panel *database.Panel, ticketId int) (uint64, error) {
	// Send welcome message
	var welcomeMessage string
	if panel == nil || panel.WelcomeMessage == nil {
		var err error
		welcomeMessage, err = dbclient.Client.WelcomeMessages.Get(guildId)
		if err != nil {
			sentry.Error(err)
			welcomeMessage = "Thank you for contacting support.\nPlease describe your issue (and provide an invite to your server if applicable) and wait for a response."
		}
	} else {
		welcomeMessage = *panel.WelcomeMessage
	}

	// %average_response%
	if isPremium && strings.Contains(welcomeMessage, "%average_response%") {
		weeklyResponseTime, err := dbclient.Client.FirstResponseTime.GetAverage(guildId, time.Hour*24*7)
		if err != nil {
			sentry.Error(err)
		} else {
			strings.Replace(welcomeMessage, "%average_response%", FormatTime(*weeklyResponseTime), -1)
		}
	}

	// variables
	welcomeMessage = doSubstitutions(welcomeMessage, worker, guildId, userId, channelId, ticketId)

	// Send welcome message
	embed := BuildEmbedRaw(worker, Green, subject, welcomeMessage, nil, isPremium)
	data := rest.CreateMessageData{
		Embed: embed,
		Components: []component.Component{
			component.BuildActionRow(
				component.BuildButton(component.Button{
					Label:    "Close",
					CustomId: "close",
					Style:    component.ButtonStyleDanger,
					Emoji:    emoji.Emoji{Name: "🔒"},
				}),
				component.BuildButton(component.Button{
					Label:    "Claim",
					CustomId: "claim",
					Style:    component.ButtonStyleSuccess,
					Emoji:    emoji.Emoji{Name: "🙋‍♂️"},
				})),
		},
	}

	msg, err := worker.CreateMessageComplex(channelId, data)
	if err != nil {
		return 0, err
	}

	return msg.Id, nil
}

func doSubstitutions(welcomeMessage string, worker *worker.Context, guildId, userId, channelId uint64, ticketId int) string {
	var lock sync.Mutex

	// do substitutions that do not require DB lookups first

	// %user%
	welcomeMessage = strings.Replace(welcomeMessage, "%user%", fmt.Sprintf("<@%d>", userId), -1)

	// %ticket_id%
	if ticketId > 0 {
		welcomeMessage = strings.Replace(welcomeMessage, "%ticket_id%", strconv.Itoa(ticketId), -1)
	}

	// %channel%
	welcomeMessage = strings.Replace(welcomeMessage, "%channel%", fmt.Sprintf("<#%d>", channelId), -1)

	// do DB lookups in parallel
	group, _ := errgroup.WithContext(context.Background())

	// %username%
	if strings.Contains(welcomeMessage, "%username%") {
		group.Go(func() (err error) {
			var user user.User
			if user, err = worker.GetUser(userId); err == nil {
				lock.Lock()
				welcomeMessage = strings.Replace(welcomeMessage, "%username%", user.Username, -1)
				lock.Unlock()
			}
			return
		})
	}

	// %server%
	if strings.Contains(welcomeMessage, "%server%") {
		group.Go(func() (err error) {
			var guild guild.Guild
			if guild, err = worker.GetGuild(guildId); err == nil {
				lock.Lock()
				welcomeMessage = strings.Replace(welcomeMessage, "%server%", guild.Name, -1)
				lock.Unlock()
			}
			return
		})
	}

	// %open_tickets%
	if strings.Contains(welcomeMessage, "%open_tickets%") {
		group.Go(func() (err error) {
			var open []database.Ticket
			if open, err = dbclient.Client.Tickets.GetGuildOpenTickets(guildId); err == nil {
				lock.Lock()
				welcomeMessage = strings.Replace(welcomeMessage, "%open_tickets%", strconv.Itoa(len(open)), -1)
				lock.Unlock()
			}
			return
		})
	}

	// %total_tickets%
	if strings.Contains(welcomeMessage, "%total_tickets%") {
		group.Go(func() (err error) {
			var count int
			if count, err = dbclient.Client.Tickets.GetTotalTicketCount(guildId); err == nil {
				lock.Lock()
				welcomeMessage = strings.Replace(welcomeMessage, "%total_tickets%", strconv.Itoa(count), -1)
				lock.Unlock()
			}
			return
		})
	}

	// %user_open_tickets%
	if strings.Contains(welcomeMessage, "%user_open_tickets%") {
		group.Go(func() (err error) {
			var open []database.Ticket
			if open, err = dbclient.Client.Tickets.GetOpenByUser(guildId, userId); err == nil {
				lock.Lock()
				welcomeMessage = strings.Replace(welcomeMessage, "%user_open_tickets%", strconv.Itoa(len(open)), -1)
				lock.Unlock()
			}
			return
		})
	}

	// %ticket_limit%
	if strings.Contains(welcomeMessage, "%ticket_limit%") {
		group.Go(func() (err error) {
			var limit uint8
			if limit, err = dbclient.Client.TicketLimit.Get(guildId); err == nil {
				lock.Lock()
				welcomeMessage = strings.Replace(welcomeMessage, "%ticket_limit%", strconv.Itoa(int(limit)), -1)
				lock.Unlock()
			}
			return
		})
	}

	if err := group.Wait(); err != nil {
		sentry.Error(err)
	}

	return welcomeMessage
}
