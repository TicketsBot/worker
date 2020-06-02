package setup

import (
	"fmt"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/message"
	"strconv"
	"strings"
)

type TicketLimitStage struct {
}

func (TicketLimitStage) State() State {
	return TicketLimit
}

func (TicketLimitStage) Prompt() string {
	return "Specify the maximum amount of tickets that a **single user** should be able to have open at once"
}

// This is not used
func (TicketLimitStage) Default() string {
	return "5"
}

func (TicketLimitStage) Process(worker *worker.Context, msg message.Message) {
	amountRaw := strings.Split(msg.Content, " ")[0]
	amount, err := strconv.Atoi(amountRaw)
	if err != nil {
		amount = 5
		utils.SendEmbed(worker, msg.ChannelId, utils.Red, "Error", fmt.Sprintf("Error: `%s`\nDefaulting to `%d`", err.Error(), amount), nil, 30, true)
		utils.ReactWithCross(worker, msg.ChannelId, msg.Id)
	} else {
		utils.ReactWithCheck(worker, msg.ChannelId, msg.Id)
	}

	if err := dbclient.Client.TicketLimit.Set(msg.GuildId, uint8(amount)); err == nil {
		utils.ReactWithCheck(worker, msg.ChannelId, msg.Id)
	} else {
		utils.ReactWithCross(worker, msg.ChannelId, msg.Id)
		sentry.Error(err)
	}
}
