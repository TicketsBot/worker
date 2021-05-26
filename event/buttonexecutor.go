package event

import (
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
	"regexp"
	"strconv"
)

var (
	panelRegex = regexp.MustCompile(`panel_(\d+)`)
)

func handleButtonPress(ctx *worker.Context, data interaction.ButtonInteraction) {
	panelRes := panelRegex.FindStringSubmatch(data.Data.CustomId)
	if len(panelRes) == 2 {
		panelId, err := strconv.Atoi(panelRes[1])
		if err != nil {
			sentry.Error(err) // TODO: Proper context
			return
		}

		panel, err := dbclient.Client.Panel.GetById(panelId)
		if err != nil {
			sentry.Error(err) // TODO: Proper context
			return
		}

		// TODO: Log this
		if panel.PanelId == 0 {
			return
		}

		// TODO: Log this
		if panel.MessageId != data.Message.Id || panel.GuildId != data.GuildId.Value {
			return
		}

		// TODO: Log this
		if data.Member == nil {
			return
		}

		// get premium tier
		premiumTier := utils.PremiumClient.GetTierByGuildId(data.GuildId.Value, true, ctx.Token, ctx.RateLimiter)
		panelCtx := command.NewPanelContext(ctx, data.GuildId.Value, data.ChannelId, data.Member.User.Id, premiumTier)

		logic.OpenTicket(&panelCtx, &panel, panel.Title)

		return
	}
}
