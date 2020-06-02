package listeners

import (
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"time"
)

func OnFirstResponse(worker *worker.Context, e *events.MessageCreate, extra eventforwarding.Extra) {
	// Make sure this is a guild
	if e.GuildId == 0 || e.Author.Bot {
		return
	}

	e.Member.User = e.Author

	// Only count replies from support reps
	if permission.GetPermissionLevel(utils.ToRetriever(worker), e.Member, e.GuildId) > 0 {
		// Retrieve ticket struct
		ticket, err := dbclient.Client.Tickets.GetByChannel(e.ChannelId)
		if err != nil {
			sentry.Error(err)
			return
		}

		// Make sure that the channel is a ticket
		if ticket.UserId != 0 {
			// We don't have to check for previous responses due to ON CONFLICT DO NOTHING
			if err := dbclient.Client.FirstResponseTime.Set(ticket.GuildId, e.Author.Id, ticket.Id, time.Now().Sub(ticket.OpenTime)); err != nil {
				sentry.Error(err)
			}
		}
	}
}

