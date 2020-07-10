package setup

import (
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/message"
)

type PrefixStage struct {
}

func (PrefixStage) State() State {
	return Prefix
}

func (PrefixStage) Prompt() translations.MessageId {
	return translations.SetupPrefix
}

func (PrefixStage) Default() string {
	return utils.DEFAULT_PREFIX
}

func (PrefixStage) Process(worker *worker.Context, msg message.Message) {
	if len(msg.Content) > 8 {
		utils.SendEmbed(worker, msg.ChannelId, msg.GuildId, utils.Red, "Error", translations.MessageInvalidPrefix, nil, 15, true, PrefixStage{}.Default())
		return
	}

	if err := dbclient.Client.Prefix.Set(msg.GuildId, msg.Content); err == nil {
		utils.ReactWithCheck(worker, msg.ChannelId, msg.Id)
	} else {
		utils.ReactWithCross(worker, msg.ChannelId, msg.Id)
		sentry.Error(err)
	}
}
