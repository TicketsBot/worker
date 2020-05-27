package setup

import (
	"fmt"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/sentry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel/message"
)

type PrefixStage struct {
}

func (PrefixStage) State() State {
	return Prefix
}

func (PrefixStage) Prompt() string {
	return "Type the prefix that you would like to use for the bot" +
		"\nThe prefix is the characters that come *before* the command (excluding the actual command itself)" +
		"\nExample: `t!`"
}

func (PrefixStage) Default() string {
	return utils.DEFAULT_PREFIX
}

func (PrefixStage) Process(worker *worker.Context, msg message.Message) {
	if len(msg.Content) > 8 {
		utils.SendEmbed(worker, msg.ChannelId, utils.Red, "Error", fmt.Sprintf("The maxium prefix langeth is 8 characters\nDefaulting to `%s`", PrefixStage{}.Default()), nil, 15, true)
		return
	}

	if err := dbclient.Client.Prefix.Set(msg.GuildId, msg.Content); err == nil {
		utils.ReactWithCheck(worker, msg.ChannelId, msg.Id)
	} else {
		utils.ReactWithCross(worker, msg.ChannelId, msg.Id)
		sentry.Error(err)
	}
}
