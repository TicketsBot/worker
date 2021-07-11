package setup

import (
	"github.com/TicketsBot/common/sentry"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/message"
)

type WelcomeMessageStage struct {
}

func (WelcomeMessageStage) State() State {
	return WelcomeMessage
}

func (WelcomeMessageStage) Prompt() database.MessageId {
	return i18n.SetupWelcomeMessage
}

func (WelcomeMessageStage) Default() string {
	return "Thank you for contacting support.\nPlease describe your issue (and provide an invite to your server if applicable) and wait for a response."
}

func (WelcomeMessageStage) Process(worker *worker.Context, msg message.Message) {
	if err := dbclient.Client.WelcomeMessages.Set(msg.GuildId, msg.Content); err == nil {
		utils.ReactWithCheck(worker, msg.ChannelId, msg.Id)
	} else {
		utils.ReactWithCross(worker, msg.ChannelId, msg.Id)
		sentry.Error(err)
	}
}
