package setup

import (
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
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

func (TicketLimitStage) Prompt() translations.MessageId {
	return translations.SetupTicketLimit
}

// This is not used
func (TicketLimitStage) Default() string {
	return "5"
}

func (TicketLimitStage) Process(worker *worker.Context, msg message.Message) {
	replyContext := utils.CreateReferenceFromMessage(msg)

	amountRaw := strings.Split(msg.Content, " ")[0]
	amount, err := strconv.Atoi(amountRaw)
	if err != nil || amount > 10 || amount < 1 {
		amount = 5
		utils.SendEmbed(worker, msg.ChannelId, msg.GuildId, replyContext, utils.Red, "Error", translations.MessageInvalidTicketLimit, nil, 30, true, amount)
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
