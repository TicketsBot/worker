package handlers

import (
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/i18n"
	"regexp"
	"strconv"
	"strings"
)

type RateHandler struct{}

func (h *RateHandler) Matcher() matcher.Matcher {
	return &matcher.FuncMatcher{
		Func: func(customId string) bool {
			return strings.HasPrefix(customId, "rate_")
		},
	}
}

func (h *RateHandler) Properties() registry.Properties {
	return registry.Properties{
		Flags: registry.SumFlags(registry.DMsAllowed, registry.CanEdit),
	}
}

var ratePattern = regexp.MustCompile(`rate_(\d+)_(\d+)_([1-5])`)

func (h *RateHandler) Execute(ctx *context.ButtonContext) {
	groups := ratePattern.FindStringSubmatch(ctx.InteractionData.CustomId)
	if len(groups) < 4 {
		return
	}

	// Errors are impossible
	guildId, _ := strconv.ParseUint(groups[1], 10, 64)
	ticketId, _ := strconv.Atoi(groups[2])
	ratingRaw, _ := strconv.Atoi(groups[3])
	rating := uint8(ratingRaw)

	// Get ticket
	ticket, err := dbclient.Client.Tickets.Get(ticketId, guildId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if ticket.UserId != ctx.InteractionUser().Id || ticket.GuildId != guildId || ticket.Id != ticketId {
		return
	}

	feedbackEnabled, err := dbclient.Client.FeedbackEnabled.Get(guildId)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if !feedbackEnabled {
		ctx.Reply(constants.Red, i18n.Error, i18n.MessageFeedbackDisabled)
		return
	}

	if err := dbclient.Client.ServiceRatings.Set(guildId, ticketId, rating); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.Reply(constants.Green, i18n.Success, i18n.MessageFeedbackSuccess)
}
