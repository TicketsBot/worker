package listeners

import (
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/setup"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

func OnSetupProgress(worker *worker.Context, e *events.MessageCreate, extra eventforwarding.Extra) {
	u := setup.SetupUser{
		Guild:   e.GuildId,
		User:    e.Author.Id,
		Channel: e.ChannelId,
		Worker:  worker,
	}

	if u.InSetup() {
		// Process current stage
		u.GetState().Process(worker, e.Message)

		// Start next stage
		u.Next()
		state := u.GetState()
		if state != nil {
			stage := state.GetStage()
			if stage != nil {
				// Psuedo-premium
				// TODO: TRANSLATE PROMPTS
				utils.SendEmbed(worker, e.ChannelId, e.GuildId, utils.Green, "Setup", (*stage).Prompt(), nil, 120, true)
			}
		}
	}
}
